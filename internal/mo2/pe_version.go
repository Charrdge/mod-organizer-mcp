package mo2

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf16"
)

// PEVersionStrings extracts FileVersion and ProductVersion from StringFileInfo in a PE DLL.
// Returns empty strings for missing or unreadable resources; err is set only on hard I/O/PE errors.
func PEVersionStrings(path string) (fileVersion, productVersion string, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	if len(raw) < 64 {
		return "", "", nil
	}
	r := bytes.NewReader(raw)
	f, err := pe.NewFile(r)
	if err != nil {
		return "", "", nil
	}
	defer f.Close()

	dd, ok := resourceDataDirectory(f)
	if !ok || dd.VirtualAddress == 0 || dd.Size == 0 {
		return "", "", nil
	}
	rootRVA := dd.VirtualAddress
	rootOff, err := rvaToFileOffset(f, rootRVA)
	if err != nil {
		return "", "", nil
	}

	verData, err := findVersionResourceRaw(raw, f, rootRVA, rootOff)
	if err != nil || len(verData) == 0 {
		return "", "", nil
	}
	fileVersion, productVersion = parseVersionStringTable(verData)
	return fileVersion, productVersion, nil
}

func resourceDataDirectory(f *pe.File) (pe.DataDirectory, bool) {
	switch h := f.OptionalHeader.(type) {
	case *pe.OptionalHeader32:
		return h.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_RESOURCE], true
	case *pe.OptionalHeader64:
		return h.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_RESOURCE], true
	default:
		return pe.DataDirectory{}, false
	}
}

func rvaToFileOffset(f *pe.File, rva uint32) (int64, error) {
	for _, s := range f.Sections {
		if rva >= s.VirtualAddress {
			end := s.VirtualAddress + s.VirtualSize
			if s.VirtualSize == 0 {
				end = s.VirtualAddress + s.Size
			}
			if rva < end {
				return int64(s.Offset) + int64(rva-s.VirtualAddress), nil
			}
		}
	}
	return 0, fmt.Errorf("unmapped rva 0x%x", rva)
}

const (
	rtVersion              = 16
	resDirFlagDirectory    = 0x80000000
	resEntryMask           = 0x7FFFFFFF
	imageResourceDataEntry = 16
)

func findVersionResourceRaw(fileRaw []byte, f *pe.File, rootRVA uint32, rootOff int64) ([]byte, error) {
	// Level 1: type — find RT_VERSION (16)
	entries1, err := readResourceDirEntries(fileRaw, rootOff)
	if err != nil {
		return nil, err
	}
	var typeOff uint32
	found := false
	for _, e := range entries1 {
		if e.nameOrID == rtVersion && (e.offset&resDirFlagDirectory) != 0 {
			typeOff = rootRVA + uint32(e.offset&resEntryMask)
			found = true
			break
		}
	}
	if !found {
		return nil, nil
	}
	typeFileOff, err := rvaToFileOffset(f, typeOff)
	if err != nil {
		return nil, err
	}
	// Level 2: name/id (usually 1)
	entries2, err := readResourceDirEntries(fileRaw, typeFileOff)
	if err != nil {
		return nil, err
	}
	var nameOff uint32
	found = false
	for _, e := range entries2 {
		if (e.offset & resDirFlagDirectory) != 0 {
			nameOff = rootRVA + uint32(e.offset&resEntryMask)
			found = true
			break
		}
	}
	if !found {
		return nil, nil
	}
	nameFileOff, err := rvaToFileOffset(f, nameOff)
	if err != nil {
		return nil, err
	}
	// Level 3: language
	entries3, err := readResourceDirEntries(fileRaw, nameFileOff)
	if err != nil {
		return nil, err
	}
	for _, e := range entries3 {
		if (e.offset & resDirFlagDirectory) != 0 {
			continue
		}
		dataStructRVA := rootRVA + uint32(e.offset&resEntryMask)
		dataStructOff, err := rvaToFileOffset(f, dataStructRVA)
		if err != nil {
			continue
		}
		if int(dataStructOff)+imageResourceDataEntry > len(fileRaw) {
			continue
		}
		dataRVA := binary.LittleEndian.Uint32(fileRaw[dataStructOff:])
		dataSize := binary.LittleEndian.Uint32(fileRaw[dataStructOff+4:])
		if dataSize == 0 || dataSize > 16*1024*1024 {
			continue
		}
		dataOff, err := rvaToFileOffset(f, dataRVA)
		if err != nil {
			continue
		}
		if dataOff < 0 || int64(dataSize) > int64(len(fileRaw))-dataOff {
			continue
		}
		return fileRaw[dataOff : dataOff+int64(dataSize)], nil
	}
	return nil, nil
}

type resDirEntry struct {
	nameOrID uint32
	offset   uint32
}

func readResourceDirEntries(fileRaw []byte, dirFileOff int64) ([]resDirEntry, error) {
	if dirFileOff < 0 || int(dirFileOff)+16 > len(fileRaw) {
		return nil, io.ErrUnexpectedEOF
	}
	off := int(dirFileOff)
	_ = binary.LittleEndian.Uint32(fileRaw[off:])   // Characteristics
	_ = binary.LittleEndian.Uint32(fileRaw[off+4:]) // TimeDateStamp
	_ = binary.LittleEndian.Uint16(fileRaw[off+8:])  // MajorVersion
	_ = binary.LittleEndian.Uint16(fileRaw[off+10:]) // MinorVersion
	named := binary.LittleEndian.Uint16(fileRaw[off+12:])
	id := binary.LittleEndian.Uint16(fileRaw[off+14:])
	total := int(named) + int(id)
	start := off + 16
	if start+total*8 > len(fileRaw) {
		return nil, io.ErrUnexpectedEOF
	}
	out := make([]resDirEntry, 0, total)
	for i := 0; i < total; i++ {
		base := start + i*8
		out = append(out, resDirEntry{
			nameOrID: binary.LittleEndian.Uint32(fileRaw[base:]),
			offset:   binary.LittleEndian.Uint32(fileRaw[base+4:]),
		})
	}
	return out, nil
}

func parseVersionStringTable(ver []byte) (fileVer, productVer string) {
	// `ver` is the raw RT_VERSION resource: root VS_VERSIONINFO, then StringFileInfo / VarFileInfo children.
	if len(ver) < 6 {
		return "", ""
	}
	key := readUTF16Z(ver, 6)
	if key != "VS_VERSIONINFO" {
		return "", ""
	}
	valLen := int(binary.LittleEndian.Uint16(ver[2:]))
	keyLen := 6 + utf16ByteLen(key)
	keyLen = alignDword(keyLen)
	childStart := 6 + keyLen + valLen
	childStart = alignDword(childStart)
	if childStart < len(ver) {
		walkStringChildren(ver[childStart:], &fileVer, &productVer)
	}
	return fileVer, productVer
}

func walkStringChildren(buf []byte, fileVer, productVer *string) {
	i := 0
	for i+6 <= len(buf) {
		length := int(binary.LittleEndian.Uint16(buf[i:]))
		if length < 6 || i+length > len(buf) {
			break
		}
		block := buf[i : i+length]
		key := readUTF16Z(block, 6)
		switch key {
		case "StringFileInfo":
			valLen := int(binary.LittleEndian.Uint16(block[2:]))
			keyLen := 6 + utf16ByteLen(key)
			keyLen = alignDword(keyLen)
			child := 6 + keyLen + valLen
			child = alignDword(child)
			if child < len(block) {
				parseStringTable(block[child:], fileVer, productVer)
			}
		case "VarFileInfo":
			// ignore
		}
		i += length
		i = alignDword(i)
	}
}

func parseStringTable(buf []byte, fileVer, productVer *string) {
	i := 0
	for i+6 <= len(buf) {
		length := int(binary.LittleEndian.Uint16(buf[i:]))
		if length < 6 || i+length > len(buf) {
			break
		}
		block := buf[i : i+length]
		// StringTable: key like "040904b0"
		_ = readUTF16Z(block, 6)
		valLen := int(binary.LittleEndian.Uint16(block[2:]))
		key := readUTF16Z(block, 6)
		keyLen := 6 + utf16ByteLen(key)
		keyLen = alignDword(keyLen)
		child := 6 + keyLen + valLen
		child = alignDword(child)
		if child < len(block) {
			parseStringPairs(block[child:], fileVer, productVer)
		}
		i += length
		i = alignDword(i)
	}
}

func parseStringPairs(buf []byte, fileVer, productVer *string) {
	i := 0
	for i+6 <= len(buf) {
		length := int(binary.LittleEndian.Uint16(buf[i:]))
		if length < 6 || i+length > len(buf) {
			break
		}
		block := buf[i : i+length]
		valLen := int(binary.LittleEndian.Uint16(block[2:]))
		key := readUTF16Z(block, 6)
		keyLen := 6 + utf16ByteLen(key)
		keyLen = alignDword(keyLen)
		valOff := 6 + keyLen
		var val string
		if valLen > 0 && valOff+valLen*2 <= len(block) {
			val = readUTF16LE(block[valOff : valOff+valLen*2])
		} else {
			val = strings.TrimSpace(readUTF16Z(block, valOff))
		}
		switch key {
		case "FileVersion":
			if *fileVer == "" {
				*fileVer = strings.TrimSpace(val)
			}
		case "ProductVersion":
			if *productVer == "" {
				*productVer = strings.TrimSpace(val)
			}
		}
		i += length
		i = alignDword(i)
	}
}

func readUTF16Z(b []byte, off int) string {
	if off >= len(b) {
		return ""
	}
	var pairs []uint16
	for i := off; i+1 < len(b); i += 2 {
		u := binary.LittleEndian.Uint16(b[i:])
		if u == 0 {
			break
		}
		pairs = append(pairs, u)
	}
	return string(utf16.Decode(pairs))
}

func readUTF16LE(b []byte) string {
	if len(b) < 2 {
		return ""
	}
	n := len(b) / 2
	pairs := make([]uint16, n)
	for i := 0; i < n; i++ {
		pairs[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	for len(pairs) > 0 && pairs[len(pairs)-1] == 0 {
		pairs = pairs[:len(pairs)-1]
	}
	return string(utf16.Decode(pairs))
}

func utf16ByteLen(s string) int {
	return (len(s) + 1) * 2
}

func alignDword(n int) int {
	return (n + 3) &^ 3
}
