# Rigol scope stats in Go

This tool allows you to pull measurement statistics from a Rigol Oscilloscope over the network connection and outputs them in CSV format. I've tested it with my DS1054Z so far.

It can also take screenshots at the same intervals. 

## Usage

    # go run .\rigol.go -host 192.168.7.90 -channels 2 -clear -interval 60 -vmax -vmin -screen > file.csv

Options are as follows. If you have more than 5 statistics displayed (on my scope at least) then reading them takes a few seconds as it has to recalculate the values as the others scroll off the list.

    -channels int
            number of channels to collect (default 4)
    -clear
            clear stats after collection
    -count int
            number of measurements to take. -1 = no limit (default -1)
    -freq
            include frequency
    -host string
            hostname or IP address (default "rigol")
    -interval int
            number of seconds between readings (default 1)
    -port int
            tcp port to use (default 5555)
    -screen
            collect screenshots in PNG format
    -vavg
            include Vavg (default true)
    -vmax
            include Vmax
    -vmin
            include Vmin
    -vpp
            include Vpp
    -vrms
            include Vrms

## Setup 

You can configure the scope over the network too - this is especially useful if you want to set specific options across a number of tests. 

I use the following example quite frequently - telnet to the scope on port 5555 and send the following commands. It's worth doing a few at a time or sometimes one will get missed if you paste in too many lines at once. 

    *RST

    :TIM:MODE ROLL
    :TIM:MAIN:SCAL 5
    :ACQ:TYPE HRES
    :DISP:TYPE DOTS

    :CHAN1:BWL ON
    :CHAN1:UNIT AMP
    :CHAN1:COUP DC
    :CHAN1:PROB 100
    :CHAN1:RANG 4
    :CHAN1:OFFS -1.5
    :CHAN1:DISP ON

    :CHAN2:BWL ON
    :CHAN2:UNIT VOLT
    :CHAN2:COUP DC
    :CHAN2:PROB 1
    :CHAN2:RANG 8
    :CHAN2:OFFS -3
    :CHAN2:DISP ON

    :MEAS:STAT:ITEM VAVG,CHAN1
    :MEAS:STAT:ITEM VAVG,CHAN2

    :CURS:MODE TRAC
    :CURS:TRAC:SOURce1 CHAN1
    :CURS:TRAC:SOURce2 CHAN2
    :CURSor:TRACk:AX 586
    :CURSor:TRACk:BX 586

Detail for these can be found in the Rigol Oscilloscope Programming guides, but in summary this sets the scope to a rolling 60s view, reading a range of -0.5 to 3.5A on Channel 1 and -1 to 7 V on channel 2, and enables the cursor for realtime measurements. 