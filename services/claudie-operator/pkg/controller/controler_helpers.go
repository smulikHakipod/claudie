/*
Copyright 2023 berops.com.

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

package controller

import (
	"fmt"

	"github.com/berops/claudie/internal/manifest"
	v1beta "github.com/berops/claudie/services/claudie-operator/pkg/api/v1beta1"
)

// mergeInputManifestWithSecrets takes the v1beta.InputManifest and providersWithSecret and returns a claudie type raw manifest.Manifest type.
// It will combine the manifest.Manifest name as object "Namespace/Name".
func mergeInputManifestWithSecrets(crd v1beta.InputManifest, providersWithSecret []v1beta.ProviderWithData, staticNodesWithSecret map[string][]v1beta.StaticNodeWithData) (manifest.Manifest, error) {
	var providers manifest.Provider

	for _, p := range providersWithSecret {
		secretNamespaceName := p.Secret.Namespace + "/" + p.Secret.Name
		switch p.ProviderType {
		case v1beta.GCP:
			gcpCredentials, err := p.GetSecretField(v1beta.GCP_CREDENTIALS)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}
			gcpProject, err := p.GetSecretField(v1beta.GCP_GCP_PROJECT)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}

			providers.GCP = append(providers.GCP, manifest.GCP{
				Name:        p.ProviderName,
				Credentials: gcpCredentials,
				GCPProject:  gcpProject,
			})

		case v1beta.AWS:
			awsAccesskey, err := p.GetSecretField(v1beta.AWS_ACCESS_KEY)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}
			awsSecretkey, err := p.GetSecretField(v1beta.AWS_SECRET_KEY)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}

			providers.AWS = append(providers.AWS, manifest.AWS{
				Name:      p.ProviderName,
				AccessKey: awsAccesskey,
				SecretKey: awsSecretkey,
			})

		case v1beta.HETZNER:
			hetzner_key, err := p.GetSecretField(v1beta.HETZNER_CREDENTIALS)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}
			var hetzner = manifest.Hetzner{
				Name:        p.ProviderName,
				Credentials: hetzner_key,
			}
			providers.Hetzner = append(providers.Hetzner, hetzner)
		case v1beta.OCI:
			ociTenant, err := p.GetSecretField(v1beta.OCI_TENANCT_OCID)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}
			ociCompartmentOcid, err := p.GetSecretField(v1beta.OCI_COMPARTMENT_OCID)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}
			ociFingerPrint, err := p.GetSecretField(v1beta.OCI_KEY_FINGERPRINT)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}
			ociPrivateKey, err := p.GetSecretField(v1beta.OCI_PRIVATE_KEY)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}
			ociUserOcid, err := p.GetSecretField(v1beta.OCI_USER_OCID)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}

			providers.OCI = append(providers.OCI, manifest.OCI{
				Name:           p.ProviderName,
				PrivateKey:     ociPrivateKey,
				KeyFingerprint: ociFingerPrint,
				TenancyOCID:    ociTenant,
				CompartmentID:  ociCompartmentOcid,
				UserOCID:       ociUserOcid,
			})
		case v1beta.AZURE:
			azureClientId, err := p.GetSecretField(v1beta.AZURE_CLIENT_ID)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}

			azureClientSecret, err := p.GetSecretField(v1beta.AZURE_CLIENT_SECRET)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}

			azureSubscriptionId, err := p.GetSecretField(v1beta.AZURE_SUBSCRIPTION_ID)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}

			azureTenantId, err := p.GetSecretField(v1beta.AZURE_TENANT_ID)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}

			providers.Azure = append(providers.Azure, manifest.Azure{
				Name:           p.ProviderName,
				SubscriptionId: azureSubscriptionId,
				TenantId:       azureTenantId,
				ClientId:       azureClientId,
				ClientSecret:   azureClientSecret,
			})
		case v1beta.CLOUDFLARE:
			cfApiToken, err := p.GetSecretField(v1beta.CF_API_TOKEN)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}
			providers.Cloudflare = append(providers.Cloudflare, manifest.Cloudflare{
				Name:     p.ProviderName,
				ApiToken: cfApiToken,
			})

		case v1beta.HETZNER_DNS:
			hetznerDNSCredentials, err := p.GetSecretField(v1beta.HETZNER_DNS_API_TOKEN)
			if err != nil {
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, err)
			}
			providers.HetznerDNS = append(providers.HetznerDNS, manifest.HetznerDNS{
				Name:     p.ProviderName,
				ApiToken: hetznerDNSCredentials,
			})
		}
	}
	// Build static nodepools
	var nodePools manifest.NodePool
	nodePools.Dynamic = crd.Spec.NodePools.Dynamic
	nodePools.Static = make([]manifest.StaticNodePool, 0, len(crd.Spec.NodePools.Static))
	// Iterate over nodepools
	for nodepool, nws := range staticNodesWithSecret {
		nodes := make([]manifest.Node, 0, len(nws))
		// Iterate over nodes and retrieve private key from secret
		for _, n := range nws {
			if key, ok := n.Secret.Data[string(v1beta.PRIVATE_KEY)]; ok {
				nodes = append(nodes, manifest.Node{Endpoint: n.Endpoint, Key: string(key)})
			} else {
				secretNamespaceName := n.Secret.Namespace + "/" + n.Secret.Name
				return manifest.Manifest{}, buildSecretError(secretNamespaceName, fmt.Errorf("field %s not found", v1beta.PRIVATE_KEY))
			}
		}
		nodePools.Static = append(nodePools.Static, manifest.StaticNodePool{Name: nodepool, Nodes: nodes})
	}

	return manifest.Manifest{
		Name:         crd.GetNamespacedNameDashed(),
		Providers:    providers,
		NodePools:    nodePools,
		Kubernetes:   crd.Spec.Kubernetes,
		LoadBalancer: crd.Spec.LoadBalancer,
	}, nil
}

// buildSecretError builds an error with the name of the NamespaceName
// of the secret, and the field in secret that is incorrect
func buildSecretError(secret string, err error) error {
	return fmt.Errorf("in secret: %s - %w", secret, err)
}

// buildProvisioningError builds a string containing errors from a single inputManifest
func buildProvisioningError(state v1beta.InputManifestStatus) error {
	var msg string
	for name, cluster := range state.Clusters {
		if cluster.State == v1beta.STATUS_ERROR {
			msg = msg + "For cluster: " + name + " Message: " + cluster.Message + "; "
		}
	}
	return fmt.Errorf(msg)
}
