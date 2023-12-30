package mechaproxy

import (
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/go-errors/errors"
)

func ensurePackage() error {
	pkgUrl := os.Getenv("PKG_URL")
	if pkgUrl == "" {
		return errors.New("PKG_URL is empty")
	}
	path := "pkg.tgz"
	if err := downloadPackge(pkgUrl, path); err != nil {
		return err
	}
	if err := exec.Command("tar", "-xf", path).Run(); err != nil {
		return errors.New(err)
	}
	return nil
}

func downloadPackge(url, path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return errors.New(err)
	}
	defer f.Close()
	resp, err := http.Get(url)
	if err != nil {
		return errors.New(err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return errors.New(err)
	}
	return nil
}
