package cloudscale

import (
	"context"
	"fmt"
	"github.com/cloudscale-ch/cloudscale-go-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func resourceCloudscaleLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudscaleLoadBalancerCreate,
		Read:   resourceCloudscaleLoadBalancerRead,
		Update: resourceCloudscaleLoadBalancerUpdate,
		Delete: resourceCloudscaleLoadBalancerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: getLoadBalancerSchema(RESOURCE),
	}
}

func getLoadBalancerSchema(t SchemaType) map[string]*schema.Schema {
	m := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Optional: t.isDataSource(),
		},
		"flavor_slug": {
			Type:     schema.TypeString,
			Required: t.isResource(),
			Computed: t.isDataSource(),
		},
		"href": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"vip_addresses": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"version": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"address": {
						Type:     schema.TypeString,
						Computed: true,
						Optional: true,
					},
					"subnet_uuid": {
						Type:     schema.TypeString,
						Computed: true,
						Optional: true,
					},
					"subnet_cidr": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"subnet_href": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
			Optional: t.isResource(),
			Computed: true,
		},
		"zone_slug": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"tags": &TagsSchema,
	}
	return m
}

func resourceCloudscaleLoadBalancerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	opts := &cloudscale.LoadBalancerRequest{
		ZonalResourceRequest: cloudscale.ZonalResourceRequest{
			Zone: d.Get("zone_slug").(string),
		},
		Name:   d.Get("name").(string),
		Flavor: d.Get("flavor_slug").(string),
	}

	vipAddressCount := d.Get("vip_addresses.#").(int)
	if vipAddressCount > 0 {
		vipAddressRequests := createVipAddressOptions(d)
		opts.VIPAddresses = &vipAddressRequests
	}

	opts.Tags = CopyTags(d)

	log.Printf("[DEBUG] LoadBalancer create configuration: %#v", opts)

	loadbalancer, err := client.LoadBalancers.Create(context.Background(), opts)
	if err != nil {
		return fmt.Errorf("Error creating LoadBalancer: %s", err)
	}

	d.SetId(loadbalancer.UUID)

	log.Printf("[INFO] LoadBalancer ID: %s", d.Id())

	fillLoadBalancerSchema(d, loadbalancer)
	return nil
}

func createVipAddressOptions(d *schema.ResourceData) []cloudscale.VIPAddressRequest {
	vipAddressCount := d.Get("vip_addresses.#").(int)
	result := make([]cloudscale.VIPAddressRequest, vipAddressCount)
	for i := 0; i < vipAddressCount; i++ {
		prefix := fmt.Sprintf("vip_addresses.%d", i)
		result[i] = cloudscale.VIPAddressRequest{
			Address: d.Get(prefix + ".address").(string),
			Subnet: cloudscale.SubnetRequest{
				UUID: d.Get(prefix + ".subnet_uuid").(string),
			},
		}
	}
	return result
}

func fillLoadBalancerSchema(d *schema.ResourceData, loadbalancer *cloudscale.LoadBalancer) {
	fillResourceData(d, gatherLoadBalancerData(loadbalancer))
}

func gatherLoadBalancerData(loadbalancer *cloudscale.LoadBalancer) ResourceDataRaw {
	m := make(map[string]interface{})
	m["id"] = loadbalancer.UUID
	m["href"] = loadbalancer.HREF
	m["name"] = loadbalancer.Name
	m["flavor_slug"] = loadbalancer.Flavor.Slug
	m["status"] = loadbalancer.Status

	if addrss := len(loadbalancer.VIPAddresses); addrss > 0 {
		vipAddressesMap := make([]map[string]interface{}, 0, addrss)
		for _, vip := range loadbalancer.VIPAddresses {

			vipMap := make(map[string]interface{})

			vipMap["version"] = vip.Version
			vipMap["address"] = vip.Address
			vipMap["subnet_uuid"] = vip.Subnet.UUID
			vipMap["subnet_cidr"] = vip.Subnet.CIDR
			vipMap["subnet_href"] = vip.Subnet.HREF

			vipAddressesMap = append(vipAddressesMap, vipMap)
		}
		m["vip_addresses"] = vipAddressesMap
	}

	return m
}

func resourceCloudscaleLoadBalancerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)

	loadbalancer, err := client.LoadBalancers.Get(context.Background(), d.Id())
	if err != nil {
		return CheckDeleted(d, err, "Error retrieving load balancer")
	}

	fillLoadBalancerSchema(d, loadbalancer)
	return nil
}

func resourceCloudscaleLoadBalancerUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	for _, attribute := range []string{"name", "flavor_slug"} {
		// cloudscale.ch Load Balancer attributes can only be changed one at a time
		if d.HasChange(attribute) {
			opts := &cloudscale.LoadBalancerRequest{}
			if attribute == "name" {
				opts.Name = d.Get(attribute).(string)
			} else if attribute == "tags" {
				opts.Tags = CopyTags(d)
			}
			err := client.LoadBalancers.Update(context.Background(), id, opts)
			if err != nil {
				return fmt.Errorf("Error updating the Load Balancer (%s): %s", id, err)
			}
		}
	}
	return resourceCloudscaleLoadBalancerRead(d, meta)
}

func resourceCloudscaleLoadBalancerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*cloudscale.Client)
	id := d.Id()

	log.Printf("[INFO] Deleting LoadBalancer: %s", id)
	err := client.LoadBalancers.Delete(context.Background(), id)

	if err != nil {
		return CheckDeleted(d, err, "Error deleting load balancer")
	}

	return nil
}
