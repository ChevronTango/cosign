/*
Copyright The Cosign Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/projectcosign/cosign/pkg/cosign"
)

func Verify() *ffcli.Command {
	var (
		flagset     = flag.NewFlagSet("cosign verify", flag.ExitOnError)
		key         = flagset.String("key", "", "path to the public key")
		checkClaims = flagset.Bool("check-claims", true, "whether to check the claims found")
		annotations = annotationsMap{}
	)
	flagset.Var(&annotations, "a", "extra key=value pairs to sign")

	return &ffcli.Command{
		Name:       "verify",
		ShortUsage: "cosign verify -key <key> <image uri>",
		ShortHelp:  "Verify a signature on the supplied container image",
		FlagSet:    flagset,
		Exec: func(ctx context.Context, args []string) error {
			if *key == "" {
				return flag.ErrHelp
			}
			if len(args) != 1 {
				return flag.ErrHelp
			}
			verified, err := VerifyCmd(ctx, *key, args[0], *checkClaims, annotations.annotations)
			if err != nil {
				return err
			}
			if !*checkClaims {
				fmt.Fprintln(os.Stderr, "Warning: the following claims have not been verified:")
			}
			for _, vp := range verified {
				fmt.Println(string(vp.Payload))
			}
			return nil
		},
	}
}

func VerifyCmd(_ context.Context, keyRef string, imageRef string, checkClaims bool, annotations map[string]string) ([]cosign.SignedPayload, error) {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return nil, err
	}

	pubKey, err := cosign.LoadPublicKey(keyRef)
	if err != nil {
		return nil, err
	}

	return cosign.Verify(ref, pubKey, checkClaims, annotations)
}
