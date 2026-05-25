package provider

import (
	"context"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

type UserDataSource struct {
	client *graphql.Client
}

type UserDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Email       types.String `tfsdk:"email"`
	Active      types.Bool   `tfsdk:"active"`
	Admin       types.Bool   `tfsdk:"admin"`
	Url         types.String `tfsdk:"url"`
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Linear user. Provide one of `id`, `name`, `email`, or `display_name` to filter.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the user.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the user.",
				Optional:            true,
				Computed:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "Display name of the user.",
				Optional:            true,
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email of the user.",
				Optional:            true,
				Computed:            true,
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the user is active.",
				Computed:            true,
			},
			"admin": schema.BoolAttribute{
				MarkdownDescription: "Whether the user is an admin.",
				Computed:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "URL of the user's profile.",
				Computed:            true,
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*graphql.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *graphql.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If ID is provided, look up directly
	if !data.Id.IsNull() && data.Id.ValueString() != "" {
		response, err := getUser(ctx, *d.client, data.Id.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
			return
		}

		data.Id = types.StringValue(response.User.Id)
		data.Name = types.StringValue(response.User.Name)
		data.DisplayName = types.StringValue(response.User.DisplayName)
		data.Email = types.StringValue(response.User.Email)
		data.Active = types.BoolValue(response.User.Active)
		data.Admin = types.BoolValue(response.User.Admin)
		data.Url = types.StringValue(response.User.Url)

		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// Otherwise, list all users and filter client-side
	response, err := getUsers(ctx, *d.client)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read users, got error: %s", err))
		return
	}

	for _, user := range response.Users.Nodes {
		if !data.Name.IsNull() && user.Name != data.Name.ValueString() {
			continue
		}
		if !data.Email.IsNull() && user.Email != data.Email.ValueString() {
			continue
		}
		if !data.DisplayName.IsNull() && user.DisplayName != data.DisplayName.ValueString() {
			continue
		}

		data.Id = types.StringValue(user.Id)
		data.Name = types.StringValue(user.Name)
		data.DisplayName = types.StringValue(user.DisplayName)
		data.Email = types.StringValue(user.Email)
		data.Active = types.BoolValue(user.Active)
		data.Admin = types.BoolValue(user.Admin)
		data.Url = types.StringValue(user.Url)

		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError("Not Found", "No user found matching the provided filters.")
}
