// Copyright 2023 Nubificus LTD.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ctr

import (
	"fmt"
	"io"
	"os"
	"time"

	gocontext "context"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/images/archive"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/pkg/epoch"
	"github.com/containerd/containerd/platforms"
)

var (
	defaultTimeout = 5 * time.Second
)

func ctrClient(address string, namespace string) (*containerd.Client, gocontext.Context, gocontext.CancelFunc, error) {
	opts := []containerd.ClientOpt{containerd.WithTimeout(defaultTimeout)}
	client, err := containerd.New(address, opts...)
	if err != nil {
		return nil, nil, nil, err
	}
	ctx, cancel := ctrAppContext(namespace)
	return client, ctx, cancel, nil
}

func ctrAppContext(namespace string) (gocontext.Context, gocontext.CancelFunc) {
	var (
		ctx    = gocontext.Background()
		cancel gocontext.CancelFunc
	)
	ctx = namespaces.WithNamespace(ctx, namespace)
	if defaultTimeout > 0 {
		ctx, cancel = gocontext.WithTimeout(ctx, defaultTimeout)
	} else {
		ctx, cancel = gocontext.WithCancel(ctx)
	}
	if tm, err := epoch.SourceDateEpoch(); err != nil {
		log.L.WithError(err).Warn("Failed to read SOURCE_DATE_EPOCH")
	} else if tm != nil {
		log.L.Debugf("Using SOURCE_DATE_EPOCH: %v", tm)
		ctx = epoch.WithSourceDateEpoch(ctx, tm)
	}
	return ctx, cancel
}

// Stripped down version of containerd/containerd/cmd/ctr/commands/images/import.go#103
func ImportImage(imageTarball string, address string, namespace string, snapshotter string) (string, error) {
	var (
		in              = imageTarball
		opts            []containerd.ImportOpt
		platformMatcher platforms.MatchComparer
	)
	client, ctx, cancel, err := ctrClient(address, namespace)
	if err != nil {
		return "", err
	}
	defer cancel()

	prefix := fmt.Sprintf("import-%s", time.Now().Format("2006-01-02"))
	opts = append(opts, containerd.WithImageRefTranslator(archive.AddRefPrefix(prefix)))

	// TODO: We could possibly skip this
	opts = append(opts, containerd.WithImportCompression())

	opts = append(opts, containerd.WithAllPlatforms(false))
	ctx, done, err := client.WithLease(ctx)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = done(ctx)
	}()

	var r io.ReadCloser
	if in == "-" {
		r = os.Stdin
	} else {
		r, err = os.Open(in)
		if err != nil {
			return "", err
		}
	}

	imgs, err := client.Import(ctx, r, opts...)
	closeErr := r.Close()
	if err != nil {
		return "", err
	}
	if closeErr != nil {
		return "", closeErr
	}

	log.G(ctx).Debugf("unpacking %d images", len(imgs))
	res := ""
	for _, img := range imgs {
		// TODO: Make sure this is correctly imported for cross-platform
		if platformMatcher == nil { // if platform not specified use default.
			platformMatcher = platforms.Default()
		}
		image := containerd.NewImageWithPlatform(client, img, platformMatcher)

		res += fmt.Sprintf("unpacked %s (%s)...", img.Name, img.Target.Digest)
		err = image.Unpack(ctx, snapshotter)
		if err != nil {
			return "", err
		}
	}
	return res, nil
}
