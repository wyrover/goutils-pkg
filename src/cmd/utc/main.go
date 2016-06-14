package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

var (
	format                   = "2006-01-02 15:04" // Format that will be used for times.
	outFormat                = format             // Output format.
	tz                       = "Local"            // String descriptor for timezone.
	fromLoc   *time.Location = time.Local         // Go time.Location for the named timezone.
	toLoc     *time.Location = time.UTC           // Go time.Location for output timezone.
)

func usage(w io.Writer) {
	fmt.Fprintf(w, `Usage: utc [-f format] [-h] [-o format] [-q] [-u] [-z zone] [time(s)...]

utc converts times to UTC. If no arguments are provided, prints the
current time in UTC. If the only time provided is "-", reads newline-
separated timestamps from standard input. If both the input and output
timezones are the same (e.g., the local time zone is UTC), a warning
message will be printed on standard error. This can be suppressed with
the -q option.

Flags:

	-f format	Go timestamp format for input times. See the Go docs
			(e.g. https://golang.org/pkg/time/#pkg-constants)
			for an explanation of this format.

			Default value: %s

	-h		Print this help message.

	-o format	Go timestamp format for outputting times. Uses the
			same format as the '-f' argument; it defaults to
			the same value as the '-f' argument.

	-q		Suppress the timezone check warning message.

	-u		Timestamps are in UTC format and should be converted
			to the timezone specified by the -z argument (which
			defaults to '%s'). Note that this isn't particularly
			useful with no arguments.

	-z zone		Text form of the time zone; this can be in short
			time zone abbreviation (e.g. MST) or a location
			(e.g. America/Los_Angeles). This has no effect when
			printing the current time.

			Default value: %s

Examples (note that the examples are done in the America/Los_Angeles /
PST8PDT time zone):

	+ Getting the current time in UTC:
	  $ utc
	  2016-06-14 14:30 = 2016-06-14 21:30
	+ Converting a local timestamp to UTC:
	  $ utc '2016-06-14 21:30'
	  2016-06-14 21:30 = 2016-06-15 04:30
	+ Converting a local EST timestamp to UTC (on a machine set to
  	  PST8PDT):
	  $ utc -z EST '2016-06-14 21:30'  
	  2016-06-14 21:30 = 2016-06-15 02:30
	+ Converting timestamps in the form '14-06-2016 3:04PM':
	  $ utc -f '02-01-2006 3:04PM' '14-06-2016 9:30PM'
	  14-06-2016 9:30PM = 15-06-2016 4:30AM
	+ Converting timestamps from standard input:
	  $ printf "2016-06-14 14:42\n2016-06-13 11:01" | utc -
	  2016-06-14 14:42 = 2016-06-14 21:42
	  2016-06-13 11:01 = 2016-06-13 18:01
	+ Converting a UTC timestamp to the local time zone:
	  $ utc -u '2016-06-14 21:30'
	  2016-06-14 21:30 = 2016-06-14 14:30
	+ Converting a UTC timestamp to EST (on a machine set to
	  PST8PDT):
	  $ utc -u -z EST '2016-06-14 21:30'
	  2016-06-14 21:30 = 2016-06-14 16:30
	+ Using a different output format:
	  $ utc -o '2006-01-02T15:03:04MST' '2016-06-14 21:30' 
	  2016-06-14 21:30 = 2016-06-15T04:04:30UTC
	+ Example of the warning message when running utc on a machine
	  where the local time zone is UTC:
	  $ utc
	   
	  ==================================================================
	  Note: both input and output timezone offsets are the same --- this
	  program may not do what you expect it to.
	   
	  (Converting from UTC (offset +0000) to UTC (offset +0000).)
	  ==================================================================
	  2016-06-14 23:44 = 2016-06-14 23:44
	+ Example of the warning message when running utc on a machine
	  where the local time zone is GMT:
	  $ utc
	   
	  ==================================================================
	  Note: both input and output timezone offsets are the same --- this
	  program may not do what you expect it to.
	   
	  (Converting from GMT (offset +0000) to UTC (offset +0000).)
	  ==================================================================
	  2016-06-14 23:46 = 2016-06-14 23:46
`, format, tz, tz)
}

func getZone(loc *time.Location) (string, int) {
	return time.Now().In(loc).Zone()
}

func checkZones(quiet bool) {
	if quiet {
		return
	}

	toZone, toOff := getZone(toLoc)
	fromZone, fromOff := getZone(fromLoc)

	if toOff == fromOff {
		fmt.Fprintf(os.Stderr, `
==================================================================
Note: both input and output timezone offsets are the same --- this
program may not do what you expect it to.

(Converting from %s (offset %+05d) to %s (offset %+05d).)
==================================================================
`, fromZone, fromOff, toZone, toOff)
	}
}

func init() {
	var help, quiet, utc bool

	flag.Usage = func() { usage(os.Stderr) }
	flag.StringVar(&format, "f", format, "time format")
	flag.BoolVar(&help, "h", false, "print usage information")
	flag.StringVar(&outFormat, "o", outFormat, "output time format")
	flag.BoolVar(&quiet, "q", false, "suppress zone check warning")
	flag.BoolVar(&utc, "u", false, "timestamps are in UTC format")
	flag.StringVar(&tz, "z", tz, "time zone to convert from; if blank, the local timezone is used")

	flag.Parse()

	if help {
		if !utc {
			checkZones(quiet)
		}
		usage(os.Stdout)
		os.Exit(0)
	}

	if utc {
		var err error
		toLoc, err = time.LoadLocation(tz)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Malformed timezone %s: %s\n", tz, err)
			os.Exit(1)
		}

		fromLoc = time.UTC
	} else {
		var err error
		fromLoc, err = time.LoadLocation(tz)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Malformed timezone %s: %s\n", tz, err)
			os.Exit(1)
		}

		if fromLoc == time.UTC {

		}

		toLoc = time.UTC
	}

	checkZones(quiet)
}

func showTime(t time.Time) {
	fmt.Printf("%s = %s\n", t.Format(format), t.In(toLoc).Format(outFormat))
}

func dumpTimes(times []string) bool {
	var errored bool

	for _, t := range times {
		u, err := time.ParseInLocation(format, t, fromLoc)
		if err != nil {
			errored = true
			fmt.Fprintf(os.Stderr, "Malformed time %s: %s\n", t, err)
			continue
		}

		showTime(u)
	}

	return errored
}

func main() {
	var times []string
	n := flag.NArg()

	switch n {
	case 0:
		showTime(time.Now())
		os.Exit(0)
	case 1:
		if flag.Arg(0) == "-" {
			s := bufio.NewScanner(os.Stdin)

			for s.Scan() {
				times = append(times, s.Text())
			}
		} else {
			times = flag.Args()
		}
	default:
		times = flag.Args()
	}

	if dumpTimes(times) {
		os.Exit(1)
	}
}
