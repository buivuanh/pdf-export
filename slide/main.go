package main

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

var (
	//go:embed html/*
	templatesFS embed.FS
	templates   *template.Template
)

func executeHTMLTemplate(path, templateName string, wr io.Writer) error {
	templates = template.Must(template.ParseGlob(fmt.Sprintf("%s/*", path)))
	err := templates.ExecuteTemplate(wr, templateName, nil)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	//templates = template.Must(
	//	template.ParseFS(templatesFS, "html/*"),
	//	//template.ParseFiles(),
	//)
	templates = template.Must(template.ParseGlob("html/*"))

}

func main() {
	go runStaticFileServer(8181)

	cliApp := &cli.App{
		Name:  "",
		Usage: "",
		Commands: []*cli.Command{
			{
				Name:    "export",
				Aliases: []string{"e"},
				Usage:   "Export pdf",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "template",
						Aliases:  []string{"t"},
						Required: true,
						Usage:    "Use file template in /path, e.g: -t=template.html",
					},
					&cli.StringFlag{
						Name:     "path",
						Aliases:  []string{"p"},
						Required: true,
						Usage:    "Set work directory, e.g: -p=html",
					},
				},
				Action: func(c *cli.Context) error {
					var eg errgroup.Group
					tplReader, tplWriter := io.Pipe()
					eg.Go(func() error {
						defer tplWriter.Close()

						err := executeHTMLTemplate(c.String("path"), c.String("template"), tplWriter)
						if err != nil {
							return err
						}

						return nil
					})

					eg.Go(func() error {
						err := exportPDF(tplReader, c.String("template"))
						if err != nil {
							return err
						}

					})

					return nil
				},
			},
		},
	}

	if err := cliApp.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Press the enter key to exit.")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
}

func exportPDF(r io.Reader, template string) error {
	err := playwright.Install()
	if err != nil {
		return err
	}

	pw, err := playwright.Run()
	if err != nil {
		return err
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(
		playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(true),
		},
	)
	if err != nil {
		return err
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		return err
	}

	buff := bytes.NewBuffer(nil)
	err = templates.ExecuteTemplate(buff, template, nil)
	if err != nil {
		return err
	}

	headerTemplate := `<div class="header" style="font-size:20px;text-align:left;padding-left:40px;width:100%;"><b>SHORTLISTED PROPERTIES</b></div>`
	footerTemplate := `<div class="footer" style="font-size:20px;text-align:center;width:100%;">Footer Content - Page <span class="pageNumber" data-page="2"></span> of <span class="totalPages"></span></div>`

	err = page.SetContent(buff.String())
	if err != nil {
		return err
	}

	//_, err = page.Goto("/github/buivuanh/pdf-export/html/template.html")
	//if err != nil {
	//	return err
	//}

	err = page.EmulateMedia(playwright.PageEmulateMediaOptions{Media: playwright.MediaPrint})
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("output/%s.pdf", time.Now().Format("20060102150405"))
	pdfOptions := playwright.PagePdfOptions{
		Path:                playwright.String(filePath),
		DisplayHeaderFooter: playwright.Bool(true),
		HeaderTemplate:      playwright.String(headerTemplate),
		FooterTemplate:      playwright.String(footerTemplate),
		PrintBackground:     playwright.Bool(true),
		Format:              playwright.String("A4"),
		//PageRanges:          playwright.String("2-"),
		Margin: &playwright.Margin{
			Top:    playwright.String("40px"),
			Bottom: playwright.String("40px"),
			Left:   playwright.String("40px"),
			Right:  playwright.String("40px"),
		},
	}

	_, err = page.PDF(pdfOptions)
	if err != nil {
		return err
	}

	fmt.Println("PDF generated successfully: ", filePath)

	return nil
}

func runStaticFileServer(port int) error {
	fs := http.FileServer(http.Dir("./html"))
	http.Handle("/", fs)

	log.Println("Listening on :8181...")
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
