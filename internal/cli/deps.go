package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/wishmatic/nano-mcp/internal/credential"
	"github.com/wishmatic/nano-mcp/internal/nanogpt"
	"github.com/wishmatic/nano-mcp/internal/resolve"
)

type Deps struct {
	In          io.Reader
	Out         io.Writer
	Err         io.Writer
	Credentials credential.Store
	Client      *nanogpt.Client
	Resolver    *resolve.Client
}

func DefaultDeps() Deps {
	return Deps{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}
}

func WithDefaults(deps Deps) Deps {
	if deps.In == nil {
		deps.In = os.Stdin
	}
	if deps.Out == nil {
		deps.Out = os.Stdout
	}
	if deps.Err == nil {
		deps.Err = os.Stderr
	}

	return deps
}

func OpenCredentials(deps Deps) (credential.Store, error) {
	if deps.Credentials != nil {
		return deps.Credentials, nil
	}

	return credential.Open()
}

func NanoClient(deps Deps) (*nanogpt.Client, error) {
	if deps.Client != nil {
		return deps.Client, nil
	}

	key := strings.TrimSpace(os.Getenv("NANOGPT_API_KEY"))
	if key == "" {
		store, err := OpenCredentials(deps)
		if err != nil {
			return nil, err
		}

		key, err = store.Load()
		if errors.Is(err, credential.ErrNotFound) {
			return nil, fmt.Errorf("not logged in: run `nano login` or set NANOGPT_API_KEY")
		}
		if err != nil {
			return nil, fmt.Errorf("load API key: %w", err)
		}
	}

	return nanogpt.New(nanogpt.Config{APIKey: key})
}
