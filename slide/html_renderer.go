package main

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/sourcegraph/conc/pool"
)

const (
	defaultMaxPages      = 16
	defaultRenderTimeout = 10 * time.Second
)

var (
	once          sync.Once
	defaultRender *Renderer

	ErrInitEngine   = errors.New("failed to init engine")
	ErrLaunchEngine = errors.New("failed to launch engine")
)

func init() {
	once.Do(func() {
		r, err := NewRenderer()
		if err != nil {
			panic(err)
		}

		defaultRender = r
	})
}

type RendererOption func(*Renderer)

func WithMaxPages(n int) RendererOption {
	return func(renderer *Renderer) {
		renderer.maxPages = n
	}
}

func WithRenderTimeout(t time.Duration) RendererOption {
	return func(renderer *Renderer) {
		renderer.pageRenderTimeout = t
	}
}

type Renderer struct {
	pw *playwright.Playwright
	hb playwright.Browser
	p  *pool.Pool

	maxPages          int
	pageRenderTimeout time.Duration
}

func NewRenderer(opts ...RendererOption) (*Renderer, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, ErrInitEngine
	}

	hb, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return nil, ErrLaunchEngine
	}

	r := &Renderer{
		pw: pw,
		hb: hb,

		maxPages:          defaultMaxPages,
		pageRenderTimeout: defaultRenderTimeout,
	}
	for _, opt := range opts {
		opt(r)
	}

	r.p = pool.New().WithMaxGoroutines(r.maxPages)
	return r, nil
}

// RenderAsImage renders the <body> tag of the given HTML document `htmlDoc` as a PNG image `pngImage`
func (r Renderer) RenderAsImage(htmlDoc io.Reader, pngImage io.Writer) error {
	errCh := make(chan error)

	// To limit number of requests at same time
	r.p.Go(func() {
		var err error
		defer func() {
			errCh <- err
		}()

		page, err := r.loadPage(htmlDoc)
		if err != nil {
			return
		}
		defer page.Close()

		elem := page.Locator("body")
		imageData, err := elem.Screenshot(playwright.LocatorScreenshotOptions{
			Type: playwright.ScreenshotTypePng,
		})
		if err != nil {
			return
		}

		_, err = io.Copy(pngImage, bytes.NewBuffer(imageData))
		if err != nil {
			return
		}
	})

	err := <-errCh
	if err != nil {
		return err
	}

	return nil
}

// RenderAsPDF renders given HTML document `htmlDoc` as a PDF
func (r Renderer) RenderAsPDF(htmlDoc io.Reader, pdf io.Writer, opts playwright.PagePdfOptions) error {
	errCh := make(chan error)

	// To limit number of requests at same time
	r.p.Go(func() {
		var err error
		defer func() {
			errCh <- err
		}()

		page, err := r.loadPage(htmlDoc)
		if err != nil {
			return
		}
		defer page.Close()

		err = page.EmulateMedia(playwright.PageEmulateMediaOptions{Media: playwright.MediaPrint})
		if err != nil {
			return
		}

		pdfData, err := page.PDF(opts)
		if err != nil {
			return
		}

		_, err = io.Copy(pdf, bytes.NewBuffer(pdfData))
		if err != nil {
			return
		}
	})

	err := <-errCh
	if err != nil {
		return err
	}

	return nil
}

func (r Renderer) loadPage(htmlDoc io.Reader) (playwright.Page, error) {
	page, err := r.hb.NewPage()
	if err != nil {
		return nil, err
	}

	htmlData, err := io.ReadAll(htmlDoc)
	if err != nil {
		return nil, err
	}

	err = page.SetContent(string(htmlData), playwright.PageSetContentOptions{
		Timeout:   playwright.Float(float64(r.pageRenderTimeout.Milliseconds())),
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	if err != nil {
		return nil, err
	}

	return page, nil
}

// RenderAsImage renders the <body> tag of the given HTML document `htmlDoc` as a PNG image `pngImage`
func RenderAsImage(htmlDoc io.Reader, pngImage io.Writer) error {
	return defaultRender.RenderAsImage(htmlDoc, pngImage)
}

func RenderAsPDF(htmlDoc io.Reader, pdf io.Writer) error {
	return defaultRender.RenderAsImage(htmlDoc, pdf)
}
