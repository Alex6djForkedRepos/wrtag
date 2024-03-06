// flagparse provides flag.Value wrappers for common types
package flagparse

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.senan.xyz/wrtag/notifications"
	"go.senan.xyz/wrtag/pathformat"
	"go.senan.xyz/wrtag/researchlink"
)

const name = "wrtag"

var (
	userConfig, _     = os.UserConfigDir()
	DefaultConfigPath = filepath.Join(userConfig, name, "config")
)

type PathFormat struct{ *pathformat.Format }

func (pf PathFormat) String() string         { return "" }
func (pf PathFormat) Set(value string) error { return pf.Parse(value) }

type Querier struct{ *researchlink.Querier }

func (q Querier) Set(value string) error {
	name, value, _ := strings.Cut(value, " ")
	name, value = strings.TrimSpace(name), strings.TrimSpace(value)
	err := q.AddSource(name, value)
	return err
}

type Notifications struct{ *notifications.Notifications }

func (n Notifications) String() string { return "" }
func (n Notifications) Set(value string) error {
	eventsRaw, uri, ok := strings.Cut(value, " ")
	if !ok {
		return fmt.Errorf("invalid notification uri format. expected \"ev1,ev2 uri\"")
	}
	var lineErrs []error
	for _, ev := range strings.Split(eventsRaw, ",") {
		ev, uri = strings.TrimSpace(ev), strings.TrimSpace(uri)
		err := n.AddURI(notifications.Event(ev), uri)
		lineErrs = append(lineErrs, err)
	}
	return errors.Join(lineErrs...)
}
