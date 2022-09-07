package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/pterm/pterm"
	"golang.org/x/tools/cover"
)

var fFunc string
var fRunTests bool
var fFileName string
var fLegend bool
var fcoverFilename string

var to = pterm.NewRGB(42, 119, 11) // This RGB value is used as the gradients start point.
var from = pterm.NewRGB(171, 200, 170)

func main() {
	flag.StringVar(&fFunc, "func", "", "Show only selected function")
	flag.StringVar(&fFileName, "file", "", "show annotated source code for selected file")
	flag.StringVar(&fcoverFilename, "coverFilename", "", "Cover profile filename location")
	flag.BoolVar(&fLegend, "legend", false, "Print sample colors")
	flag.BoolVar(&fRunTests, "runtests", true, "Run tests and generate coverage profile")
	flag.Parse()

	// set fFileName or fFunc if this is given as argument
	// i.e gocover main.go
	setFlagsFromArgs()

	if fcoverFilename == "" {
		outFile, err := ioutil.TempFile("", "coverage")
		if err != nil {
			log.Fatal(err)
		}
		fcoverFilename = outFile.Name()
		defer func() {
			err := os.Remove(fcoverFilename)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	if fLegend {
		legend(os.Stdout)
		os.Exit(0)
	}

	if fRunTests {
		cmd := exec.Command(
			"go",
			"test",
			"-covermode=count",
			fmt.Sprintf("-coverprofile=%s", fcoverFilename),
			"./...",
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(string(output))
			log.Fatal(err)
		}
	}

	profiles, err := cover.ParseProfiles(fcoverFilename)
	if err != nil {
		log.Fatal(err)
	}

	if fFunc != "" {
		err := printfunc(fFunc, profiles)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return

	}
	if fFileName != "" {
		printFile(profiles, fFileName)
		return
	}
	generateReport(profiles)
}

// set fFileName or fFunc if this is given as argument
// i.e gocover main.go
func setFlagsFromArgs() {
	args := flag.Args()
	if len(args) == 1 {
		if strings.HasSuffix(args[0], ".go") && fileExists(args[0]) {
			fFileName = args[0]
		} else {
			// asume a function name was given
			fFunc = args[0]
		}
	}
}

// printfunc finds funcname and prints it with coverage head map
func printfunc(funcname string, profiles []*cover.Profile) error {
	for _, profile := range profiles {
		fn := profile.FileName

		actualfile, err := findFile(fn)
		if err != nil {
			return err
		}

		funcs, err := findFuncs(actualfile)
		if err != nil {
			return err
		}

		for _, f := range funcs {
			if f.name == funcname {
				analyzeandprintWithFunc(actualfile, profile, f)
				return nil
			}
		}
	}
	return fmt.Errorf("could not find function %s", funcname)
}

// run through cover profiles and print coverage of files and functions
func generateReport(profiles []*cover.Profile) {
	for _, profile := range profiles {

		actualfile, err := findFile(profile.FileName)
		if err != nil {
			log.Fatal(err)
		}
		// print file info
		printFileAndCoverage(actualfile, percentCovered(profile))
		// print function info
		printFunctionsAndCoverage(profile, actualfile)

	}

	cov := totalcoverage(profiles)
	fmt.Printf(`-------------------------------------------------
Total covered: %s
`, fadeprint(fmt.Sprintf("%.2f%%", cov), cov))
}

// find filename and print the contents coverage head map
func printFile(profiles []*cover.Profile, filename string) {
	for _, profile := range profiles {

		actualfile, err := findFile(profile.FileName)
		if err != nil {
			log.Fatal(err)
		}
		// print the file
		if len(fFileName) > 0 && strings.Contains(actualfile, filename) {
			analyzeandprint(actualfile, profile)
		}
	}
}

func analyzeandprint(actualfile string, profile *cover.Profile) {
	src, err := ioutil.ReadFile(actualfile)
	if err != nil {
		log.Printf("can't read %q: %v", profile.FileName, err)
		os.Exit(0)
	}
	paintpoints, err := generatePaintPoints(src, profile.Boundaries(src))
	if err != nil {
		log.Fatal(err)
	}
	printCoverage(paintpoints, src, 0, len(src))
}

func analyzeandprintWithFunc(actualfile string, profile *cover.Profile, fun *FuncExtent) {
	src, err := ioutil.ReadFile(actualfile)
	if err != nil {
		log.Printf("can't read %q: %v", profile.FileName, err)
		os.Exit(0)
	}
	pp, err := generatePaintPoints(src, profile.Boundaries(src))
	if err != nil {
		log.Fatal(err)
	}
	start, stop := findstartstop(src, fun)
	printCoverage(pp, src, start, stop)
}

func findstartstop(s []byte, fun *FuncExtent) (start, stop int) {
	line := 1
	posinline := 0

	for i, v := range s {
		posinline++

		if line == fun.startLine && posinline == fun.startCol {
			start = i
		}
		if line == fun.endLine && posinline == fun.endCol {
			stop = i
			return
		}
		if v == '\n' {
			line++
			posinline = 0
		}
	}
	log.Fatal("could not find start/stop")
	return
}

// find functions in file and print names and coverage
func printFunctionsAndCoverage(profile *cover.Profile, file string) error {
	funcs, err := findFuncs(file)
	if err != nil {
		return err
	}

	tabber := tabwriter.NewWriter(os.Stdout, 1, 8, 1, '\t', 0)
	defer tabber.Flush()
	var total, covered int64

	// Now match up functions and profile blocks.
	for _, f := range funcs {
		c, t := f.coverage(profile)
		cov := 100.0 * float64(c) / float64(t)
		fmt.Fprintf(tabber, "%s:%d:\t%s\t%s\n",
			path.Base(file),
			f.startLine,
			f.name,
			fadeprint(fmt.Sprintf("%.1f%%", cov), cov))
		total += t
		covered += c
	}
	fmt.Fprint(tabber, "\n")
	return nil
}

func fadeprint(s string, cov float64) string {

	if cov == 0 {
		return pterm.FgRed.Sprint(s)
	}

	if cov > 99 {
		return to.Sprint(pterm.NewStyle(pterm.Bold).Sprint(s))
	}

	return from.Fade(0, float32(10), float32(int(cov/10)), to).Sprint(s)

}
func printFileAndCoverage(filename string, cov float64) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fn := strings.Replace(filename, fmt.Sprintf("%s/", wd), "", 1)
	pterm.NewStyle(pterm.Bold).Printf("# %-20s ", fn)
	fmt.Print(fadeprint(fmt.Sprintf("%.1f%%\n", cov), cov))
	fmt.Println("---------------------------")
}
func totalcoverage(profiles []*cover.Profile) float64 {
	var total, covered int64
	for _, p := range profiles {
		for _, b := range p.Blocks {
			total += int64(b.NumStmt)
			if b.Count > 0 {
				covered += int64(b.NumStmt)
			}
		}
	}
	if total == 0 {
		return 0
	}
	return float64(covered) / float64(total) * 100
}

func percentCovered(p *cover.Profile) float64 {
	var total, covered int64
	for _, b := range p.Blocks {
		total += int64(b.NumStmt)
		if b.Count > 0 {
			covered += int64(b.NumStmt)
		}
	}
	if total == 0 {
		return 0
	}
	return float64(covered) / float64(total) * 100
}

func printCoverage(pp []paintpoint, src []byte, start, stop int) {
	pi := 0 // paintpoint index
	var w bytes.Buffer
	for i := start; i < stop; i++ {
		if pi < len(pp) && i >= pp[pi].start && i <= pp[pi].stop {

			fmt.Fprint(&w, fadeprint(string(src[i]), float64(pp[pi].cov)))
		} else {
			fmt.Fprint(&w, pterm.FgLightWhite.Sprint(string(src[i])))
		}
		if pi < len(pp) && i >= pp[pi].stop {
			pi++
		}

	}
	fmt.Println(w.String())
}

// paintpoint represent a chunk of code with start/stop byte index in
// sourcefile, coverage and statement count
type paintpoint struct {
	cov   int
	start int
	stop  int
	count int
}

func generatePaintPoints(src []byte, boundaries []cover.Boundary) ([]paintpoint, error) {
	paintboard := make([]paintpoint, 0)

	pp := paintpoint{}
	for i := range src {
		for len(boundaries) > 0 && boundaries[0].Offset == i {
			b := boundaries[0]
			if b.Start {
				n := 0
				if b.Count > 0 {
					n = int(math.Floor(b.Norm*99)) + 1
				}

				pp.start = i
				pp.cov = n
				pp.count = b.Count
			} else {
				pp.stop = i
				paintboard = append(paintboard, pp)
				pp = paintpoint{}
			}
			boundaries = boundaries[1:]
		}

	}

	return paintboard, nil
}
