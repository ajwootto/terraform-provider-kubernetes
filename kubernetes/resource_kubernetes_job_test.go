package kubernetes

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	api "k8s.io/client-go/pkg/apis/batch/v1"
)

func TestAccKubernetesJob_basic(t *testing.T) {
	var conf api.Job
	name := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "kubernetes_job.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckKubernetesJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKubernetesJobConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesJobExists("kubernetes_job.test", &conf),
					resource.TestCheckResourceAttr("kubernetes_job.test", "metadata.0.name", name),
					resource.TestCheckResourceAttrSet("kubernetes_job.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("kubernetes_job.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("kubernetes_job.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("kubernetes_job.test", "metadata.0.uid"),
					resource.TestCheckResourceAttr("kubernetes_job.test", "spec.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_job.test", "spec.0.parallelism", "2"),
					resource.TestCheckResourceAttr("kubernetes_job.test", "spec.0.template.0.spec.0.container.0.name", "hello"),
				),
			},
			{
				Config: testAccKubernetesJobConfig_modified(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKubernetesJobExists("kubernetes_job.test", &conf),
					resource.TestCheckResourceAttr("kubernetes_job.test", "metadata.0.name", name),
					resource.TestCheckResourceAttrSet("kubernetes_job.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("kubernetes_job.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("kubernetes_job.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("kubernetes_job.test", "metadata.0.uid"),
					resource.TestCheckResourceAttr("kubernetes_job.test", "spec.#", "1"),
					resource.TestCheckResourceAttr("kubernetes_job.test", "spec.0.parallelism", "2"),
					resource.TestCheckResourceAttr("kubernetes_job.test", "spec.0.template.0.spec.0.container.0.name", "hello"),
				),
			},
		},
	})
}

func testAccCheckKubernetesJobDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*kubernetes.Clientset)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "kubernetes_job" {
			continue
		}

		namespace, name, err := idParts(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := conn.BatchV1().Jobs(namespace).Get(name, meta_v1.GetOptions{})
		if err == nil {
			if resp.Name == rs.Primary.ID {
				return fmt.Errorf("Job still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckKubernetesJobExists(n string, obj *api.Job) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*kubernetes.Clientset)

		namespace, name, err := idParts(rs.Primary.ID)
		if err != nil {
			return err
		}

		out, err := conn.BatchV1().Jobs(namespace).Get(name, meta_v1.GetOptions{})
		if err != nil {
			return err
		}

		*obj = *out
		return nil
	}
}

func testAccKubernetesJobConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "kubernetes_job" "test" {
	metadata {
		name = "%s"
	}
	spec {
		parallelism = 2
		template {
			metadata {
				labels {
					job = "one"
				}
			}
			spec {
				container {
					name = "hello"
					image = "alpine"
					command = ["echo", "'hello'"]
				}
			}
		}
	}
}`, name)
}

func testAccKubernetesJobConfig_modified(name string) string {
	return fmt.Sprintf(`
resource "kubernetes_job" "test" {
	metadata {
		name = "%s"
	}
	spec {
		parallelism = 2
		template {
			spec {
				container {
					name = "hello"
					image = "alpine"
					command = ["echo", "'world'"]
				}
			}
		}
	}
}`, name)
}
