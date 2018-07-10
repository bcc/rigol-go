
package main

import (
	"fmt"
	"net"
	"bufio"
	"time"
	"os"
	"flag"
	"strings"
	"strconv"
)

var chancount = flag.Int("channels", 4, "number of channels to collect")
var host = flag.String("host", "rigol", "hostname or IP address")
var port = flag.Int("port", 5555, "tcp port to use")
var interval = flag.Int("interval", 1, "number of seconds between readings")
var count = flag.Int("count", -1, "number of measurements to take. -1 = no limit")
var f_vavg = flag.Bool("vavg", true, "include Vavg")
var f_vmin = flag.Bool("vmin", true, "include Vmin")
var f_vmax = flag.Bool("vmax", true, "include Vmax")
var f_vpp = flag.Bool("vpp", false, "include Vpp")
var f_vrms = flag.Bool("vrms", false, "include Vrms")
var f_freq = flag.Bool("freq", false, "include frequency")
var f_screen = flag.Bool("screen", false, "collect screenshots in PNG format")

func main() {

	flag.Parse()

	connstr := fmt.Sprintf("%s:%d", *host, *port)
	conn, err := net.Dial("tcp", connstr)
	if err != nil {
		fmt.Printf("Unable to connect to %s\n", connstr)
		os.Exit(2)
	}

	toRun, header := buildQuery()

	// Output CSV header
	fmt.Printf("%s, %s, %s\n", "timestamp", header, "querytime")

	for ; *count != 0; *count-- {

		tstart := time.Now()

		// Collect and write screenshot if the screen flag is set.
		if *f_screen {
			img := getScreenshot(conn)
			writeScreenshot(img)
		}

		result := queryScope(conn, toRun)
		// Return from scope is semicolon separated, so we'll just switch for commas.
		result = strings.Replace(result, ";", ", ", -1)

		tdone := time.Now()
		fmt.Printf("%s, %s, %s\n", tdone.Format(time.RFC3339), result, tdone.Sub(tstart))

		time.Sleep(time.Duration(*interval)*time.Second)
	}
}

func queryScope (conn net.Conn, query string) string {
	fmt.Fprintf(conn, fmt.Sprintf("%s\n", query))
	c1, _ := bufio.NewReader(conn).ReadString('\n')
	return strings.TrimSuffix(c1, "\n")
}

func getScreenshot (conn net.Conn) []byte {
	// Args are Colour, Invert, Format.
	fmt.Fprintf(conn, ":DISP:DATA? ON,FALSE,PNG\n")
	reader := bufio.NewReader(conn)
	header1,_ := reader.ReadByte()

	// should be a #
	if header1 != 35 {
		return nil
	}

	// Next character shows us the length of the length of the datastream!
	header2,_ := reader.ReadByte()
	buffsize := int(header2 - 48)

	header3 := make([]byte, buffsize)
	for i:=0; i<buffsize; i++ {
		t,_ := reader.ReadByte()
		header3[i] = t
	}
 
	// This is now the image buffersize
	buffsize,_ = strconv.Atoi(string(header3))

	imgdata := make([]byte, buffsize)
	for i:=0; i<buffsize; i++ {
		t,_ := reader.ReadByte()
		imgdata[i] = t
	}

	return imgdata
}

func writeScreenshot (img []byte) {
	// Filename safe date format
	tstamp := time.Now().Format("2006-01-02.150405")
	filename := fmt.Sprintf("%s.%s.png", "screenshot", tstamp)
		
	f, err := os.Create(filename)
	check(err)
	defer f.Close()
	_, err = f.Write(img)
	check(err)
}

func buildQuery () (string,string) {
	var command []string
	var header []string
	for i:=0; i<*chancount; i++ {

		// Stacking queries together seems to knock about 10% off the query time
		// compared to requesting one measurement at a time.
		// 2 channels, 4 measurements resulted in 3.9 vs 3.5 seconds on my DS1054Z.

		// Tell the scope which channel the following commands are for
		command = append(command, fmt.Sprintf(":MEAS:SOUR CHAN%d", i+1))

		// Query measurement unit.
		command = append(command, fmt.Sprintf(":CHAN%d:UNIT?", i+1))
		header = append(header, fmt.Sprintf("CH%d Unit", i+1))

		if *f_vavg {
			command = append(command, ":MEAS:ITEM? VAVG")
			header = append(header, fmt.Sprintf("CH%d Vavg", i+1))
		}
		if *f_vmin {
			command = append(command, ":MEAS:ITEM? VMIN")
			header = append(header, fmt.Sprintf("CH%d Vmin", i+1))
		}
		if *f_vmax {
			command = append(command, ":MEAS:ITEM? VMAX")
			header = append(header, fmt.Sprintf("CH%d Vmax", i+1))
		}
		if *f_vpp {
			command = append(command, ":MEAS:ITEM? VPP")
			header = append(header, fmt.Sprintf("CH%d Vpp", i+1))
		}
		if *f_vrms {
			command = append(command, ":MEAS:ITEM? VRMS")
			header = append(header, fmt.Sprintf("CH%d Vrms", i+1))
		}
		if *f_freq {
			command = append(command, ":MEAS:ITEM? FREQ")
			header = append(header, fmt.Sprintf("CH%d freq", i+1))
		}
	}
	return strings.Join(command,";"), strings.Join(header,", ")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}