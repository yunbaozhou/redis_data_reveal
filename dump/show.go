package dump

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/cli"
	"github.com/xueqiu/rdr/decoder"
	"github.com/xueqiu/rdr/static"
)

var counters = NewSafeMap()

func listPathFiles(pathname string) []string {
	var filenames []string
	fi, err := os.Lstat(pathname) // For read access.
	if err != nil {
		return filenames
	}
	if fi.IsDir() {
		files, err := ioutil.ReadDir(pathname)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			name := path.Join(pathname, f.Name())
			filenames = append(filenames, name)
		}
	} else {
		filenames = append(filenames, pathname)
	}
	return filenames
}

// Show parse rdbfile(s) and show statistical information by html
func Show(c *cli.Context) {
	if c.NArg() < 1 {
		fmt.Fprintln(c.App.ErrWriter, "show requires at least 1 argument")
		cli.ShowCommandHelp(c, "show")
		return
	}

	// Initialize history manager
	InitHistoryManager("history.json")

	// parse rdbfile
	fmt.Fprintln(c.App.Writer, "start parsing...")
	instances := []string{}
	InitHTMLTmpl()
	go func() {
		for {
			for _, pathname := range c.Args() {
				for _, v := range listPathFiles(pathname) {
					filename := filepath.Base(v)

					if !counters.Check(filename) {
						decoder := decoder.NewDecoder()
						fmt.Fprintf(c.App.Writer, "start to parse %v \n", filename)
						go Decode(c, decoder, v)
						counter := NewCounter()
						counter.Count(decoder.Entries)
						counters.Set(filename, counter)
						fmt.Fprintf(c.App.Writer, "parse %v  done\n", filename)

						instances = append(instances, filename)
						// init html template
						// init common data in template
						tplCommonData["Instances"] = instances

						// Save to history
						fileInfo, _ := os.Stat(v)
						hm := GetHistoryManager()

						// Calculate totals
						var totalKeys, totalBytes uint64
						for _, v := range counter.typeNum {
							totalKeys += v
						}
						for _, v := range counter.typeBytes {
							totalBytes += v
						}

						historyEntry := HistoryEntry{
							Filename:    filename,
							FilePath:    v,
							UploadTime:  time.Now(),
							FileSize:    fileInfo.Size(),
							TotalKeys:   totalKeys,
							TotalMemory: totalBytes,
						}
						hm.Add(historyEntry)
					}
				}

			}
			time.Sleep(5 * time.Second)
		}
	}()

	// start http server
	startHTTPServer(c, instances)
}

// ShowWeb starts the web server with upload capability
func ShowWeb(c *cli.Context) {
	// Initialize history manager
	InitHistoryManager("history.json")

	// Load instances from history
	instances := []string{}
	hm := GetHistoryManager()
	historyEntries := hm.GetAll()
	for _, entry := range historyEntries {
		instances = append(instances, entry.Filename)
	}

	InitHTMLTmpl()
	tplCommonData["Instances"] = instances

	// start http server
	startHTTPServer(c, instances)
}

func startHTTPServer(c *cli.Context, instances []string) {
	staticFS := assetfs.AssetFS{
		Asset:     static.Asset,
		AssetDir:  static.AssetDir,
		AssetInfo: static.AssetInfo,
	}
	router := httprouter.New()
	router.ServeFiles("/static/*filepath", &staticFS)
	router.GET("/", showMainPage)
	router.GET("/instance/:path", rdbReveal)
	router.GET("/terminal/:path", showTerminal)
	router.POST("/api/upload", uploadHandler)
	router.GET("/list", listInstances)
	router.GET("/api/progress/:path", progressHandler)
	router.GET("/api/stream/:path", streamLogsHandler)
	router.GET("/api/history", historyHandler)

	// Ops analysis endpoints
	router.GET("/api/ops/analysis/:path", opsAnalysisHandler)
	router.GET("/api/ops/anomalies/:path", opsAnomaliesHandler)
	router.GET("/api/ops/recommendations/:path", opsRecommendationsHandler)
	router.GET("/api/ops/health/:path", opsHealthHandler)

	// Create HTTP server with custom timeouts for large file uploads
	server := &http.Server{
		Addr:              ":" + c.String("port"),
		Handler:           router,
		ReadTimeout:       30 * time.Minute, // 30 minutes for reading large files
		ReadHeaderTimeout: 1 * time.Minute,
		WriteTimeout:      30 * time.Minute, // 30 minutes for writing response
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	fmt.Fprintln(c.App.Writer, "Server started, please access http://localhost:"+c.String("port"))
	fmt.Fprintln(c.App.Writer, "You can upload RDB files or access existing instances")
	fmt.Fprintln(c.App.Writer, "Note: Large file uploads may take several minutes")

	listenErr := server.ListenAndServe()
	if listenErr != nil {
		fmt.Fprintf(c.App.ErrWriter, "Listen port err: %v\n", listenErr)
	}
}

// showMainPage renders the main page with upload capability
func showMainPage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "views/main_with_upload.html")
}

// listInstances returns a list of available instances
func listInstances(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	instances := []string{}
	for key := range counters.Items() {
		instances = append(instances, key.(string))
	}
	response := map[string]interface{}{
		"instances": instances,
	}
	json.NewEncoder(w).Encode(response)
}

// historyHandler returns the analysis history
func historyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	hm := GetHistoryManager()
	entries := hm.GetAll()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": entries,
	})
}
