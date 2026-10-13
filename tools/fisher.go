package tools

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var Fisher = &BasicTool{
	ToolName:        "Fisher",
	ToolCmd:         "fisher",
	Shell:           "fish",
	ToolDescription: "A plugin manager for Fish",
	UpdateArgs:      []string{"update"},
	VersionArgs:     []string{"--version"},
	ListOutdated:    listFisherOutdated,
}

// fisherPluginsScript prints one line per installed plugin: its name and the files fisher installed for it, tab separated.
const fisherPluginsScript = `
for plugin in $_fisher_plugins
    set files _fisher_(string escape --style=var -- $plugin)_files
    string join \t $plugin (string replace -- \~ ~ $$files)
end
`

// listFisherOutdated reports the plugins whose installed files differ from their repository on GitHub,
// since fisher keeps no record of what it installed.
func listFisherOutdated(ctx context.Context) ([]Package, error) {
	output, err := exec.CommandContext(ctx, "fish", "-c", fisherPluginsScript).Output()
	if err != nil {
		return nil, err
	}

	var packages []Package
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Split(line, "\t")
		name, files := fields[0], fields[1:]

		// Only a GitHub plugin, "owner/repo" with an optional "@ref", can be checked.
		// Fisher also takes a local path or a gitlab.com repository.
		repository, ref, pinned := strings.Cut(name, "@")
		if strings.Count(repository, "/") != 1 {
			continue
		}
		if !pinned {
			ref = "HEAD"
		}

		tree, err := githubTree(ctx, repository, ref)
		if err != nil {
			return nil, err
		}
		installed, err := installedFiles(files)
		if err != nil {
			return nil, err
		}
		if !maps.Equal(installed, upstreamFiles(tree)) {
			packages = append(packages, Package{Name: name})
		}
	}
	return packages, nil
}

// installedFiles hashes a plugin's installed files the way git does, keyed by their path in the repository.
// Fisher installs each file straight into one of its four directories, so that is the file's last two components.
func installedFiles(files []string) (map[string]string, error) {
	hashes := map[string]string{}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		hashes[filepath.Base(filepath.Dir(file))+"/"+filepath.Base(file)] = gitBlobHash(content)
	}
	return hashes, nil
}

func gitBlobHash(content []byte) string {
	hash := sha1.New()
	fmt.Fprintf(hash, "blob %d\x00", len(content))
	hash.Write(content)
	return hex.EncodeToString(hash.Sum(nil))
}

type githubTreeEntry struct {
	Path string `json:"path"`
	SHA  string `json:"sha"`
}

// githubTree lists the repository's files and directories at ref, each with its git hash.
func githubTree(ctx context.Context, repository, ref string) ([]githubTreeEntry, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/git/trees/%s?recursive=1", repository, ref)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, response.Status)
	}

	var tree struct {
		Tree []githubTreeEntry `json:"tree"`
	}
	if err := json.NewDecoder(response.Body).Decode(&tree); err != nil {
		return nil, err
	}
	return tree.Tree, nil
}

// upstreamFiles returns the files fisher installs from the repository, keyed by path:
// those in its functions, conf.d, completions, and themes directories.
func upstreamFiles(tree []githubTreeEntry) map[string]string {
	files := map[string]string{}
	for _, entry := range tree {
		switch path.Dir(entry.Path) {
		case "functions", "conf.d", "completions", "themes":
			files[entry.Path] = entry.SHA
		}
	}
	return files
}
