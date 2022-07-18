// Copyright 2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hacbs-contract/ec-cli/internal/image"
)

func signOffCmd() *cobra.Command {
	var data = struct {
		imageRef  string
		publicKey string
	}{
		imageRef:  "",
		publicKey: "",
	}
	cmd := &cobra.Command{
		Use:   "sign-off",
		Short: "Capture signed off signatures from a source (github repo, Jira)",
		Long: `Supported sign off sources are commits captured from a git repo and jira issues.
               The git sources return a signed off value and the git commit. The jira issue is
			   a TODO, but will return the Jira issue with any sign off values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			imageValidator, err := image.NewImageValidator(cmd.Context(), data.imageRef, data.publicKey, "")
			if err != nil {
				return err
			}

			validatedImage, err := imageValidator.ValidateImage(cmd.Context())
			if err != nil {
				return err
			}

			for _, att := range validatedImage.Attestations {
				signoffSource, err := att.NewSignOffSource()
				if err != nil {
					return err
				}
				if signoffSource == nil {
					return errors.New("there is no signoff source in attestation")
				}

				signOff, err := signoffSource.GetSignOff()
				if err != nil {
					return err
				}

				if signOff != nil {
					payload, err := json.Marshal(signOff)
					if err != nil {
						return err
					}
					fmt.Println(string(payload))
				}
			}
			return nil
		},
	}

	// attestation download options
	cmd.Flags().StringVar(&data.publicKey, "public-key", "", "Public key")
	cmd.Flags().StringVar(&data.imageRef, "image-ref", data.imageRef, "The OCI repo to fetch the attestation from.")

	return cmd
}

func init() {
	rootCmd.AddCommand(signOffCmd())
}
