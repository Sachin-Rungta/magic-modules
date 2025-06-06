package acctest

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Hardcode the Terraform resource name -> API service name mapping temporarily.
// TODO: [tgc] read the mapping from the resource metadata files.
var ApiServiceNames = map[string]string{
	"google_compute_instance": "compute.googleapis.com",
	"google_project":          "cloudresourcemanager.googleapis.com",
}

// Gets the test metadata for tgc:
//   - test config
//   - cai asset name
//     For example: //compute.googleapis.com/projects/ci-test-188019/zones/us-central1-a/instances/tf-test-mi3fqaucf8
func GetTestMetadataForTgc(service, address, rawConfig string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		splits := strings.Split(address, ".")
		if splits == nil || len(splits) < 2 {
			return fmt.Errorf("The resource address %s is invalid.", address)
		}
		resourceType := splits[0]
		resourceName := splits[1]

		rState := s.RootModule().Resources[address]
		if rState == nil || rState.Primary == nil {
			return fmt.Errorf("The resource state is unavailable. Please check if the address %s.%s is correct.", resourceType, resourceName)
		}

		// Convert the resource ID into CAI asset name
		// and then print out the CAI asset name in the logs
		if apiServiceName, ok := ApiServiceNames[resourceType]; !ok {
			return fmt.Errorf("The Cai product backend name for resource %s doesn't exist.", resourceType)
		} else {
			var rName string
			switch resourceType {
			case "google_project":
				rName = fmt.Sprintf("projects/%s", rState.Primary.Attributes["number"])
			default:
				rName = rState.Primary.ID
			}
			caiAssetName := fmt.Sprintf("//%s/%s", apiServiceName, rName)
			log.Printf("[DEBUG]TGC CAI asset names start\n%s\nEnd of TGC CAI asset names", caiAssetName)
		}

		// The acceptance tests names will be also used for the tgc tests.
		// "service" is logged and will be used to put the tgc tests into specific service packages.
		log.Printf("[DEBUG]TGC Terraform service: %s", service)
		log.Printf("[DEBUG]TGC Terraform resource: %s", address)

		re := regexp.MustCompile(`\"(tf[-_]?test[-_]?.*?)([a-z0-9]+)\"`)
		rawConfig = re.ReplaceAllString(rawConfig, `"${1}tgc"`)
		log.Printf("[DEBUG]TGC raw_config starts %sEnd of TGC raw_config", rawConfig)
		return nil
	}
}
