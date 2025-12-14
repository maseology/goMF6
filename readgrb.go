package gomf6

import (
	"bytes"
	"encoding/binary"
	"log"
	"strconv"
	"strings"

	"github.com/maseology/goMF6/readers"
	"github.com/maseology/mmaths"
	"github.com/maseology/mmio"
)

func readGRB(fp string) ([]*mmaths.Prism, [][]int, []readers.JAxr, []int, []int) {
	buf := mmio.OpenBinary(fp)
	var btyp, bver [50]byte
	if err := binary.Read(buf, binary.LittleEndian, &btyp); err != nil {
		log.Fatalln("Fatal error: readGRB read 001 failed: ", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &bver); err != nil {
		log.Fatalln("Fatal error: readGRB read 002 failed: ", err)
	}
	ttyp, tver := strings.TrimSpace(string(btyp[:])), strings.TrimSpace(string(bver[:]))
	if tver != "VERSION 1" {
		log.Fatalf("Error:\n GRB %s version not supported: '%s'", fp, tver)
	}

	switch ttyp {
	case "GRID DIS":
		readGRBasciiHeader(buf)
		return readers.ReadGRBgrid(buf)
	case "GRID DISU":
		readGRBasciiHeader(buf)
		return readers.ReadGRBU(buf)
	default:
		log.Fatalf("GRB type '%s' currently not supported", ttyp)
		return nil, nil, nil, nil, nil
	}
}

func readGRBasciiHeader(b *bytes.Reader) {
	// read past *.grb ascii header
	var bntxt, blentxt [50]byte
	if err := binary.Read(b, binary.LittleEndian, &bntxt); err != nil {
		log.Fatalln("Fatal error: readGRBheader read 001 failed: ", err)
	}
	if err := binary.Read(b, binary.LittleEndian, &blentxt); err != nil {
		log.Fatalln("Fatal error: readGRBheader read 002 failed: ", err)
	}
	ntxt, err := strconv.Atoi(strings.TrimSpace(string(bntxt[:])[5:]))
	if err != nil {
		log.Fatalln("Fatal error: readGRBheader read 003 failed: ", err)
	}
	lentxt, err := strconv.Atoi(strings.TrimSpace(string(blentxt[:])[7:]))
	if err != nil {
		log.Fatalln("Fatal error: readGRBheader read 004 failed: ", err)
	}
	for {
		ln := make([]byte, lentxt)
		if err := binary.Read(b, binary.LittleEndian, ln); err != nil {
			log.Fatalln("Fatal error: readGRBheader read 005 failed: ", err)
		}
		// fmt.Println(strings.TrimSpace(string(ln[:])))
		ntxt--
		if ntxt <= 0 {
			break
		}
	}
}
