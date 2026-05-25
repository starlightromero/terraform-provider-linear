package provider

import (
	"context"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UsersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

type UsersDataSource struct {
	client *graphql.Client
}

type UsersDataSourceUserModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Email       types.String `tfsdk:"email"`
	Active      types.Bool   `tfsdk:"active"`
	Admin       types.Bool   `tfsdk:"admin"`
	Url         types.String `tfsdk:"url"`
}

type UsersDataSourceModel struct {
	Users []UsersDataSourceUserModel `tfsdk:"users"`
}

func (d *UsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Linear users.",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				MarkdownDescription: "List of users.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Identifier of the user.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the user.",
							Computed:            true,
						},
						"display_name": schema.StringAttribute{
							MarkdownDescription: "Display name of the user.",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "Email of the user.",
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
				},
			},
		},
	}
}

func (d *UsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := getUsers(ctx, *d.client)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read users, got error: %s", err))
		return
	}

	for _, user := range response.Users.Nodes {
		data.Users = append(data.Users, UsersDataSourceUserModel{
			Id:          types.StringValue(user.Id),
			Name:        types.StringValue(user.Name),
			DisplayName: types.StringValue(user.DisplayName),
			Email:       types.StringValue(user.Email),
			Active:      types.BoolValue(user.Active),
			Admin:       types.BoolValue(user.Admin),
			Url:         types.StringValue(user.Url),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
