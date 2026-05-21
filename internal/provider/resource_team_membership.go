package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &TeamMembershipResource{}
var _ resource.ResourceWithImportState = &TeamMembershipResource{}

func NewTeamMembershipResource() resource.Resource {
	return &TeamMembershipResource{}
}

type TeamMembershipResource struct {
	client *graphql.Client
}

type TeamMembershipResourceModel struct {
	Id     types.String `tfsdk:"id"`
	TeamId types.String `tfsdk:"team_id"`
	UserId types.String `tfsdk:"user_id"`
	Owner  types.Bool   `tfsdk:"owner"`
}

func (r *TeamMembershipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_membership"
}

func (r *TeamMembershipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Linear team membership.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the team membership.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team_id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the team.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex(), "must be an uuid"),
				},
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the user.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex(), "must be an uuid"),
				},
			},
			"owner": schema.BoolAttribute{
				MarkdownDescription: "Whether the user is a team owner.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *TeamMembershipResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*graphql.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *graphql.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *TeamMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *TeamMembershipResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := TeamMembershipCreateInput{
		TeamId: data.TeamId.ValueString(),
		UserId: data.UserId.ValueString(),
		Owner:  data.Owner.ValueBool(),
	}

	response, err := createTeamMembership(ctx, *r.client, input)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team membership, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "created a team membership")

	membership := response.TeamMembershipCreate.TeamMembership

	data.Id = types.StringValue(membership.Id)
	data.TeamId = types.StringValue(membership.Team.Id)
	data.UserId = types.StringValue(membership.User.Id)
	data.Owner = types.BoolValue(membership.Owner)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TeamMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *TeamMembershipResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := getTeamMembership(ctx, *r.client, data.Id.ValueString())

	if err != nil {
		if strings.Contains(err.Error(), "Entity not found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team membership, got error: %s", err))
		return
	}

	membership := response.TeamMembership

	data.Id = types.StringValue(membership.Id)
	data.TeamId = types.StringValue(membership.Team.Id)
	data.UserId = types.StringValue(membership.User.Id)
	data.Owner = types.BoolValue(membership.Owner)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TeamMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *TeamMembershipResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	input := TeamMembershipUpdateInput{
		Owner: data.Owner.ValueBool(),
	}

	response, err := updateTeamMembership(ctx, *r.client, input, data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update team membership, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "updated a team membership")

	membership := response.TeamMembershipUpdate.TeamMembership

	data.Id = types.StringValue(membership.Id)
	data.TeamId = types.StringValue(membership.Team.Id)
	data.UserId = types.StringValue(membership.User.Id)
	data.Owner = types.BoolValue(membership.Owner)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TeamMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *TeamMembershipResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := deleteTeamMembership(ctx, *r.client, data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete team membership, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted a team membership")
}

func (r *TeamMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
