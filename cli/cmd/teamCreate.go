package cmd

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/util/errors"
	"github.com/argoproj/argo-cd/v2/util/localconfig"
	"github.com/spf13/cobra"
)

// teamCreateCmd represents the teamCreate command
var teamCreateCmd = &cobra.Command{
	Use:   "create  <team name> optional: -p <parent team name> -s <path to pipeline schemas>",
	Short: "Command to create a team.",
	Long: `
Command to create a team. Specify the name of the team to be created.
The optional -p flag is used to set the parent team name, and is 'na' by default. The
filename of a JSON file with defined pipeline schemas is also optional and set with the -s flag. 
If provided, the created team will automatically have these pipelines defined.
	 
Example usage:
	atlas team create team_name (team will be created under 'na' parent team by default)
	atlas team create team_name -p parent_team
	atlas team create team_name -p parent_team -s pipeline_schemas.json`,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Invalid number of arguments. Run atlas team create -h to see usage details.")
			return
		}
		defaultLocalConfigPath, err := localconfig.DefaultLocalConfigPath()
		errors.CheckError(err)
		config, _ := localconfig.ReadLocalConfig(defaultLocalConfigPath)
		context, _ := config.ResolveContext(apiclient.ClientOptions{}.Context)

		parentTeamName, _ := cmd.Flags().GetString("parent")
		teamName := args[0]

		url := "https://" + atlasURL + "/team/" + orgName + "/" + parentTeamName + "/" + teamName

		var req *http.Request
		var er error

		if cmd.Flags().Lookup("schemas").Changed {
			jsonFile, err := os.Open(args[3])
			if err != nil {
				fmt.Println("Unable to find or process pipeline schemas file")
			}
			defer jsonFile.Close()

			byteValue, _ := ioutil.ReadAll(jsonFile)
			req, er = http.NewRequest("POST", url, bytes.NewReader(byteValue))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", context.User.AuthToken))
			if er != nil {
				fmt.Println("Request failed, please try again.")
			}
		} else {
			req, er = http.NewRequest("POST", url, nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", context.User.AuthToken))
			if er != nil {
				fmt.Println("Request failed, please try again.")
			}
		}

		certPEM := `-----BEGIN CERTIFICATE-----
MIIFOzCCAyOgAwIBAgIQRxTE4jdMRNoFXqG1Z6u0uzANBgkqhkiG9w0BAQsFADAm
MQswCQYDVQQGEwJVUzEXMBUGA1UEChMOR3JlZW5PcHMsIElOQy4wHhcNMjExMjMw
MjEwNTU5WhcNMzExMjMwMjEwNTU5WjAmMQswCQYDVQQGEwJVUzEXMBUGA1UEChMO
R3JlZW5PcHMsIElOQy4wggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQDJ
7cg1TSP/czsAq2Ca5ItndyaB6azQn8l0oVkXbT858E5XnFxtedzsmgu6+lNcOEon
1kKf3Z8HvMkw0rGlNI0MHIIB/FunB79LIonUmgBCNb+PXJcaCmsFVoRFXNKMjh89
4QpjGhy1XFw1yLfTUQpYsScx9OdhgZFihkga+/RD6Nr3oadr0n6omZmn+eAfvaGm
99xkY+sy47vtCRDFGJyCb37fDZZC1R914TBhcIAMeVMCUnIQe49xTLwDF4CSGVB7
RE8rwZ3KtPM0tso+ZpyJXjVyvPboDIRYn7DJld7sVV1Atf039DLfhmviANw3BO6m
r3iVcGOyhF+D9nfkqW3jHSHdgpBEMjbKs+kqzlnVERb/8bFH0GGA5oK5f5nb8urP
aBr11T6OEZRrfMAD4BzAP6XR+GGWISQsOs9KvZDSRKmYHvPlhvJKeoREqVHjHisE
lHctxumLJf1FS5oPFfDVd0Gg447Okc0Ne9eE0ajBM9o/Yr2xaAOvyfyQa/nyj0GV
cgoTFLCDUJmV9AADEH48dA3w8N0j2MuvCtkOALdL6v4aaHE/wL3cd77p1hDgrgkh
ZfHfKHtsUCIUivc+pWJknPiHczGRQzgdkbQj+mKJygJ4AdY/lWqR2JcneFETbbmr
ZspYR8mm141plfI37Z3eMUnWYAPHw4W+GbKTCUEgTQIDAQABo2UwYzAOBgNVHQ8B
Af8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDgYDVR0OBAcEBQECAwQGMCwG
A1UdEQQlMCOCCWxvY2FsaG9zdIcEfwAAAYcQAAAAAAAAAAAAAAAAAAAAATANBgkq
hkiG9w0BAQsFAAOCAgEANQ1i34RTA9N/ZoKvpGqcTZktOgehBu28ppFhFDkpsHv4
G3Sx/ls9Eg11m+TnyAjDA1G8lEJeO6P+br+BmrAj4cVYcRRfTJrRA8t1zfpKVaks
wnqEuxc1iOpecOAXD5rG/yuibCgsT+ZRE6d+GZVT2Iv4tjoq4KW3UdWRvDOuk9kr
k7q6v84jx20V0LRtKH5pnsdkpNqDkm+3Mgu9MeASGpINuX7OL3sbaHlZxs44wSFW
0mRW8ehroKGms9r8Z9KD//mun9HTgyhEE4pVIG9eghBKimfvibIdZTa39m6EL5E2
xhEejiNJun3HMe1mVW/FEqtR8zCb7vSoyi1uTCaJYmQR4rZ1oj9Ku4fGsgEOx6KS
FwlcCx8/b5BsHeDtMs8ef/mXIvR0GwxxO2U6Dj6i0B0t+Binzp2skw48uqh9Fj6Z
eXjkzKhtS1/Pp6AzADZROgM71BzhOekswO3alNmz9DO3jAGCALu8PB5Y09WHgbtG
XU9MfdIvFj6wB6mUcOCFjTbC7WtfFwE5d/T4wXL5MhF25Dbluq8FcIAVnclack5t
J2q79aptXZekQaWEiGtBjBglf7B4PDwP4kWRrVnRT4A5eJacg26g5AoHLt4Jn8RD
lWiSBervciIiV63WDR5PmLabXPZ5EWK6qu1BD0erG9Z534wYF4XmWb3rnZDKMoY=
-----END CERTIFICATE-----`

		// TODO: seems like http client should be configured with TLS config and shared between all functions
		// TODO: currently we are adding PEM certificate to the pool
		//		but maybe this certificate should be added by workflowtrigger
		//		need to investigate more deeply
		//		in a case we cannot reach root ca pool the certificate should be added to pool on each side
		//		as in this case we will create a different pool on each side
		certpool := bestEffortSystemCertPool()
		certpool.AppendCertsFromPEM([]byte(certPEM))
		clientTLSConf := &tls.Config{
			RootCAs: certpool,
		}
		transport := &http.Transport{
			TLSClientConfig: clientTLSConf,
		}
		client := &http.Client{Timeout: 20 * time.Second, Transport: transport}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Request failed with the following error:", err)
			return
		}
		statusCode := resp.StatusCode
		if statusCode == 200 {
			fmt.Println("Successfully created team:", teamName, "under parent team:", parentTeamName)
		} else if statusCode == 400 {
			fmt.Println("Team creation failed because the request was invalid.\nPlease check if org and parent team names are correct, a team with the specified name doesn't already exist, and the format of the schema file (if provided) is valid.")
		} else {
			fmt.Println("Internal server error: ", err)
		}
	},
}

// TODO: shared function
func bestEffortSystemCertPool() *x509.CertPool {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		log.Println("root ca not found, returning new...")
		return x509.NewCertPool()
	}
	log.Println("root ca found")
	return rootCAs
}

func init() {
	teamCmd.AddCommand(teamCreateCmd)
	teamCreateCmd.PersistentFlags().StringP("parent", "p", "na", "parent team name")
	teamCreateCmd.PersistentFlags().StringP("schemas", "s", "", "path to pipeline schemas JSON file")

}
