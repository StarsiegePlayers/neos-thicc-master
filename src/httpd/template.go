package httpd

import (
	"fmt"
	"io"
	"text/template"
)

type htmlTemplate struct {
	Hostname string
}

func (rt *Router) InitTemplate() (err error) {
	index, err := rt.embedFS.Open("index.html")
	if err != nil {
		return fmt.Errorf("unable to open index.html [%w]", err)
	}

	indexBytes, err := io.ReadAll(index)
	if err != nil {
		return fmt.Errorf("unable to read index.html [%w]", err)
	}

	rt.indexTemplate, err = template.New("index_html").Parse(string(indexBytes))
	if err != nil {
		return fmt.Errorf("unable to parse index.html as a template [%w]", err)
	}

	return
}
