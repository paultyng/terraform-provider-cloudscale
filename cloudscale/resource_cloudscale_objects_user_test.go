package cloudscale

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/cloudscale-ch/cloudscale-go-sdk/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("cloudscale_objects_user", &resource.Sweeper{
		Name: "cloudscale_objects_user",
		F:    testSweepObjectsUsers,
	})
}

func testSweepObjectsUsers(region string) error {
	meta, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client := meta.(*cloudscale.Client)

	ObjectsUsers, err := client.ObjectsUsers.List(context.Background())
	if err != nil {
		return err
	}

	foundError := error(nil)
	for _, u := range ObjectsUsers {
		if strings.HasPrefix(u.DisplayName, "terraform-") {
			log.Printf("Destroying ObjectsUser %#v", u.DisplayName)

			if err := client.ObjectsUsers.Delete(context.Background(), u.ID); err != nil {
				foundError = err
			}
		}
	}
	return foundError
}

func TestAccCloudscaleObjectsUser_Minimal(t *testing.T) {
	var objectsUser cloudscale.ObjectsUser

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleObjectsUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: objectsUserConfigMinimal(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleObjectsUserExists("cloudscale_objects_user.basic", &objectsUser),
					resource.TestCheckResourceAttr("cloudscale_objects_user.basic", "display_name", fmt.Sprintf("terraform-%d", rInt)),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "href"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "user_id"),
					resource.TestCheckResourceAttr("cloudscale_objects_user.basic", "keys.#", "1"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "keys.0.access_key"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "keys.0.secret_key"),
				),
			},
		},
	})
}

func TestAccCloudscaleObjectsUser_Rename(t *testing.T) {
	var objectsUser cloudscale.ObjectsUser

	rInt1 := acctest.RandInt()
	rInt2 := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleObjectsUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: objectsUserConfigMinimal(rInt1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleObjectsUserExists("cloudscale_objects_user.basic", &objectsUser),
					resource.TestCheckResourceAttr("cloudscale_objects_user.basic", "display_name", fmt.Sprintf("terraform-%d", rInt1)),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "href"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "user_id"),
					resource.TestCheckResourceAttr("cloudscale_objects_user.basic", "keys.#", "1"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "keys.0.access_key"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "keys.0.secret_key"),
				),
			},
			{
				Config: objectsUserConfigMinimal(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleObjectsUserExists("cloudscale_objects_user.basic", &objectsUser),
					resource.TestCheckResourceAttr("cloudscale_objects_user.basic", "display_name", fmt.Sprintf("terraform-%d", rInt2)),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "href"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "user_id"),
					resource.TestCheckResourceAttr("cloudscale_objects_user.basic", "keys.#", "1"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "keys.0.access_key"),
					resource.TestCheckResourceAttrSet("cloudscale_objects_user.basic", "keys.0.secret_key"),
				),
			},
		},
	})
}

func TestAccCloudscaleObjectsUser_import_basic(t *testing.T) {
	var afterImport, afterUpdate cloudscale.ObjectsUser

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleObjectsUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: objectsUserConfigMinimal(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleObjectsUserExists("cloudscale_objects_user.basic", &afterImport),
					resource.TestCheckResourceAttr(
						"cloudscale_objects_user.basic", "display_name", fmt.Sprintf("terraform-%d", rInt)),
				),
			},
			{
				ResourceName:      "cloudscale_objects_user.basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "cloudscale_objects_user.basic",
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateId:     "does-not-exist",
				ExpectError:       regexp.MustCompile(`Cannot import non-existent remote object`),
			},
			{
				Config: objectsUserConfigMinimal(42),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudscaleObjectsUserExists("cloudscale_objects_user.basic", &afterUpdate),
					resource.TestCheckResourceAttr(
						"cloudscale_objects_user.basic", "display_name", "terraform-42"),
					testAccCheckObjectsUserIsSame(t, &afterImport, &afterUpdate),
				),
			},
		},
	})
}

func TestAccCloudscaleObjectsUser_tags(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudscaleObjectsUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: objectsUserConfigWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_objects_user.basic", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_objects_user.basic", "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						"cloudscale_objects_user.basic", "tags.my-bar", "bar"),
					testTagsMatch("cloudscale_objects_user.basic"),
				),
			},
			{
				Config: objectsUserConfigMinimal(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_objects_user.basic", "tags.%", "0"),
					testTagsMatch("cloudscale_objects_user.basic"),
				),
			},
			{
				Config: objectsUserConfigWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"cloudscale_objects_user.basic", "tags.%", "2"),
					resource.TestCheckResourceAttr(
						"cloudscale_objects_user.basic", "tags.my-foo", "foo"),
					resource.TestCheckResourceAttr(
						"cloudscale_objects_user.basic", "tags.my-bar", "bar"),
					testTagsMatch("cloudscale_objects_user.basic"),
				),
			},
		},
	})
}

func testAccCheckObjectsUserIsSame(t *testing.T,
	before, after *cloudscale.ObjectsUser) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if adr := before; adr == after {
			t.Fatalf("Passed the same instance twice, address is equal=%v",
				adr)
		}
		if before.ID != after.ID {
			t.Fatalf("Not expected a change of Objects User IDs got=%s, expected=%s",
				after.ID, before.ID)
		}
		return nil
	}
}

func testAccCheckCloudscaleObjectsUserDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*cloudscale.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudscale_objects_user" {
			continue
		}

		id := rs.Primary.ID

		// Try to find the objectsUser
		v, err := client.ObjectsUsers.Get(context.Background(), id)
		if err == nil {
			return fmt.Errorf("objects user %v still exists", v)
		} else {
			errorResponse, ok := err.(*cloudscale.ErrorResponse)
			if !ok || errorResponse.StatusCode != http.StatusNotFound {
				return fmt.Errorf(
					"Error waiting for objectsUser (%s) to be destroyed: %s",
					rs.Primary.ID, err)
			}
		}
	}

	return nil
}

func objectsUserConfigMinimal(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_objects_user" "basic" {
  display_name    = "terraform-%d"
}
`, rInt)
}

func objectsUserConfigWithTags(rInt int) string {
	return fmt.Sprintf(`
resource "cloudscale_objects_user" "basic" {
  display_name    = "terraform-%d"
  tags = {
    my-foo = "foo"
    my-bar = "bar"
  }
}
`, rInt)
}
