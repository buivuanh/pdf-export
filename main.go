package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/urfave/cli/v2"
)

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
						Usage:    "Use file template in /path/template, e.g: -t=template.html",
					},
					&cli.StringFlag{
						Name:     "path",
						Aliases:  []string{"p"},
						Required: true,
						Usage:    "Set work directory, e.g: -p=html",
					},
					&cli.StringFlag{
						Name:    "header",
						Aliases: []string{"H"},
						Value:   "header.html",
						Usage:   "Set print header in /path/header, default: header.html",
					},
					&cli.StringFlag{
						Name:    "footer",
						Aliases: []string{"f"},
						Value:   "footer.html",
						Usage:   "Set print footer in /path/footer, default: footer.html",
					},
					&cli.StringFlag{
						Name:    "data",
						Aliases: []string{"d"},
						Usage:   "Use json as data source for template in /path/data",
					},
				},
				Action: func(c *cli.Context) error {
					htmlData, headerData, footerData, err := executeHTMLTemplate(
						c.String("path"),
						c.String("template"),
						c.String("header"),
						c.String("footer"),
						c.String("data"),
					)
					if err != nil {
						return err
					}

					err = exportPDF(htmlData, headerData, footerData)
					if err != nil {
						return err
					}

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

func exportPDF(htmlData, headerData, footerData string) error {
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

	err = page.SetContent(htmlData)
	if err != nil {
		return err
	}

	err = page.EmulateMedia(playwright.PageEmulateMediaOptions{Media: playwright.MediaPrint})
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("output/%s.pdf", time.Now().Format("20060102150405"))
	pdfOptions := playwright.PagePdfOptions{
		Path:                playwright.String(filePath),
		DisplayHeaderFooter: playwright.Bool(true),
		HeaderTemplate:      playwright.String(headerData),
		FooterTemplate:      playwright.String(footerData),
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

	dir, _ := os.Getwd()
	fmt.Println("PDF generated successfully: ", filePath)
	fmt.Printf("Path: %s/%s \n", dir, filePath)
	fmt.Printf("Open: file://%s/%s \n", dir, filePath)

	return nil
}

func readData(path string) (map[string]any, error) {
	// Read JSON file
	jsonFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse JSON data
	var data map[string]any
	err = json.Unmarshal(jsonFile, &data)
	if err != nil {
		log.Fatalf("Error parsing JSON data: %v", err)
	}

	return data, nil
}

func executeHTMLTemplate(path, templateName, headerFile, footerFile, dataFile string) (htmlData, headerData, footerData string, err error) {
	var pageData map[string]any
	if dataFile != "" {
		pageData, err = readData(fmt.Sprintf("%s/%s", path, dataFile))
		if err != nil {
			return "", "", "", err
		}
	}

	templates := template.Must(template.ParseGlob(fmt.Sprintf("%s/*", path)))

	html := bytes.NewBuffer(nil)
	err = templates.ExecuteTemplate(html, templateName, pageData)
	if err != nil {
		return "", "", "", err
	}

	header := bytes.NewBuffer(nil)
	err = templates.ExecuteTemplate(header, headerFile, nil)
	if err != nil {
		return "", "", "", err
	}

	footer := bytes.NewBuffer(nil)
	err = templates.ExecuteTemplate(footer, footerFile, nil)
	if err != nil {
		return "", "", "", err

	}

	return html.String(), header.String(), footer.String(), nil
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
