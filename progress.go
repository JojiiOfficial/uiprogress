package uiprogress

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/gosuri/uilive"
)

// Out is the default writer to render progress bars to
var Out = os.Stdout

// RefreshInterval in the default time duration to wait for refreshing the output
var RefreshInterval = time.Millisecond * 10

// defaultProgress is the default progress
var defaultProgress = New()

// Progress represents the container that renders progress bars
type Progress struct {
	// Out is the writer to render progress bars to
	Out io.Writer

	// Width is the width of the progress bars
	Width int

	// Bars is the collection of progress bars
	Bars []*Bar

	// RefreshInterval in the time duration to wait for refreshing the output
	RefreshInterval time.Duration

	lw     *uilive.Writer
	ticker *time.Ticker
	tdone  chan bool
	mtx    *sync.RWMutex
}

// New returns a new progress bar with defaults
func New() *Progress {
	lw := uilive.New()
	lw.Out = Out

	return &Progress{
		Width:           Width,
		Out:             Out,
		Bars:            make([]*Bar, 0),
		RefreshInterval: RefreshInterval,

		tdone: make(chan bool),
		lw:    uilive.New(),
		mtx:   &sync.RWMutex{},
	}
}

// AddBar adds bar to the default progress container
func AddBar(bar *Bar) *Bar {
	return defaultProgress.AddBar(bar)
}

// AddNewBar creates a new progress bar and adds it to the default progress container
func AddNewBar(total int) *Bar {
	return defaultProgress.AddNewBar(total)
}

// Start starts the rendering the progress of progress bars using the DefaultProgress. It listens for updates using `bar.Set(n)` and new bars when added using `AddBar`
func Start() {
	defaultProgress.Start()
}

// Stop stops listening
func Stop() {
	defaultProgress.Stop()
}

// Listen listens for updates and renders the progress bars
func Listen() {
	defaultProgress.Listen()
}

// SetOut sets output
func (p *Progress) SetOut(o io.Writer) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	p.Out = o
	p.lw.Out = o
}

// SetRefreshInterval sets refresh interval
func (p *Progress) SetRefreshInterval(interval time.Duration) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.RefreshInterval = interval
}

// AddNewBar creates a new progress bar and adds to the container
func (p *Progress) AddNewBar(total int) *Bar {
	bar := NewBar(total)
	return p.AddBar(bar)
}

// AddBar creates a new progress bar and adds to the container
func (p *Progress) AddBar(bar *Bar) *Bar {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	bar.Width = p.Width
	p.Bars = append(p.Bars, bar)
	return bar
}

// Listen listens for updates and renders the progress bars
func (p *Progress) Listen() {
	for {

		p.mtx.Lock()
		interval := p.RefreshInterval
		p.mtx.Unlock()

		select {
		case <-time.After(interval):
			p.print()
		case <-p.tdone:
			p.print()
			close(p.tdone)
			return
		}
	}
}

func (p *Progress) print() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	for _, bar := range p.Bars {
		fmt.Fprintln(p.lw, bar.String())
	}
	p.lw.Flush()
}

// Start starts the rendering the progress of progress bars. It listens for updates using `bar.Set(n)` and new bars when added using `AddBar`
func (p *Progress) Start() {
	go p.Listen()
}

// Stop stops listening
func (p *Progress) Stop() {
	p.tdone <- true
	<-p.tdone
}

// Bypass returns a writer which allows non-buffered data to be written to the underlying output
func (p *Progress) Bypass() io.Writer {
	return p.lw.Bypass()
}
