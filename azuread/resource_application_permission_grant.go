package azuread

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/terraform-providers/terraform-provider-azuread/azuread/helpers/p"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azuread/azuread/helpers/validate"
)

const resourceOAuth2PermissionGrant = "azuread_application_permission_grant"

func resourcePermissionGrant() *schema.Resource {
	return &schema.Resource{
		Create: resourcePermissionGrantCreate,
		Read:   resourcePermissionGrantRead,
		Update: resourcePermissionGrantUpdate,
		Delete: resourcePermissionGrantDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.UUID,
			},

			"object_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.UUID,
			},

			"resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"consent_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"AllPrincipal",
					"Principal",
				}, false),
			},

			"scope": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"start_time": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"expiry_time": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"grant_time": {
				Type:         schema.TypeString,
				Computed:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},
		},
	}
}

func resourcePermissionGrantCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).oauth2PermissionGrantClient
	ctx := meta.(*ArmClient).StopContext

	// Required parameters
	clientID := d.Get("client_id").(string)
	objectID := d.Get("object_id").(string)
	resourceID := d.Get("resource_id").(string)
	consentType := d.Get("consent_type").(graphrbac.ConsentType)

	grant := &graphrbac.OAuth2PermissionGrant{
		ClientID:    p.String(clientID),
		ObjectID:    p.String(objectID),
		ResourceID:  p.String(resourceID),
		ConsentType: consentType,
	}

	// Optional
	if v, ok := d.GetOk("scope"); ok {
		grant.Scope = p.String(v.(string))
	}

	timeNow := time.Now()

	if v, ok := d.GetOk("start_time"); ok {
		grant.StartTime = p.String(v.(string))
	} else {
		grant.StartTime = p.String(timeNow.String())
	}

	if v, ok := d.GetOk("expiry_time"); ok {
		grant.ExpiryTime = p.String(v.(string))
	} else {
		expiryTime := timeNow.AddDate(2, 0, 0)
		grant.ExpiryTime = p.String(expiryTime.String())
	}

	resp, err := client.Create(ctx, grant)

	if err != nil {
		return fmt.Errorf("Error creating permission grant: %+v", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to creating permission grant: %+v", err)
	}

	return resourcePermissionGrantRead(d, meta)
}

func resourcePermissionGrantRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).oauth2PermissionGrantClient
	ctx := meta.(*ArmClient).StopContext

	objectID := d.Get("object_id").(string)
	resp, err := client.List(ctx, objectID)

	if err != nil {
		return fmt.Errorf("Failed to retrieve permission grant for the application %q : %+v", objectID, err)
	}

	if resp.Response().StatusCode != 200 {
		return fmt.Errorf("Failed to creating permission grant: %+v", err)
	}

	return nil
}

func resourcePermissionGrantUpdate(d *schema.ResourceData, meta interface{}) error {
	// Delete old grant
	resourcePermissionGrantDelete(d, meta)

	// Create new grant
	return resourcePermissionGrantCreate(d, meta)
}

func resourcePermissionGrantDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).oauth2PermissionGrantClient
	ctx := meta.(*ArmClient).StopContext

	objectID := d.Get("object_id").(string)

	resp, err := client.Delete(ctx, objectID)

	if err != nil {
		return fmt.Errorf("Failed to delete existing permission grant %q : %+v", objectID, err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to creating permission grant: %+v", err)
	}

	return nil
}
