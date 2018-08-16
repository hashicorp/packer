package clientconfig

// PublicClouds represents a collection of PublicCloud entries in clouds-public.yaml file.
// The format of the clouds-public.yml is documented at
// https://docs.openstack.org/python-openstackclient/latest/configuration/
type PublicClouds struct {
	Clouds map[string]Cloud `yaml:"public-clouds"`
}

// Clouds represents a collection of Cloud entries in a clouds.yaml file.
// The format of clouds.yaml is documented at
// https://docs.openstack.org/os-client-config/latest/user/configuration.html.
type Clouds struct {
	Clouds map[string]Cloud `yaml:"clouds"`
}

// Cloud represents an entry in a clouds.yaml/public-clouds.yaml/secure.yaml file.
type Cloud struct {
	Cloud      string        `yaml:"cloud"`
	Profile    string        `yaml:"profile"`
	AuthInfo   *AuthInfo     `yaml:"auth"`
	AuthType   AuthType      `yaml:"auth_type"`
	RegionName string        `yaml:"region_name"`
	Regions    []interface{} `yaml:"regions"`

	// API Version overrides.
	IdentityAPIVersion string `yaml:"identity_api_version"`
	VolumeAPIVersion   string `yaml:"volume_api_version"`

	// Verify whether or not SSL API requests should be verified.
	Verify *bool `yaml:"verify"`

	// CACertFile a path to a CA Cert bundle that can be used as part of
	// verifying SSL API requests.
	CACertFile string `yaml:"cacert"`

	// ClientCertFile a path to a client certificate to use as part of the SSL
	// transaction.
	ClientCertFile string `yaml:"cert"`

	// ClientKeyFile a path to a client key to use as part of the SSL
	// transaction.
	ClientKeyFile string `yaml:"key"`
}

// AuthInfo represents the auth section of a cloud entry or
// auth options entered explicitly in ClientOpts.
type AuthInfo struct {
	// AuthURL is the keystone/identity endpoint URL.
	AuthURL string `yaml:"auth_url"`

	// Token is a pre-generated authentication token.
	Token string `yaml:"token"`

	// Username is the username of the user.
	Username string `yaml:"username"`

	// UserID is the unique ID of a user.
	UserID string `yaml:"user_id"`

	// Password is the password of the user.
	Password string `yaml:"password"`

	// ProjectName is the common/human-readable name of a project.
	// Users can be scoped to a project.
	// ProjectName on its own is not enough to ensure a unique scope. It must
	// also be combined with either a ProjectDomainName or ProjectDomainID.
	// ProjectName cannot be combined with ProjectID in a scope.
	ProjectName string `yaml:"project_name"`

	// ProjectID is the unique ID of a project.
	// It can be used to scope a user to a specific project.
	ProjectID string `yaml:"project_id"`

	// UserDomainName is the name of the domain where a user resides.
	// It is used to identify the source domain of a user.
	UserDomainName string `yaml:"user_domain_name"`

	// UserDomainID is the unique ID of the domain where a user resides.
	// It is used to identify the source domain of a user.
	UserDomainID string `yaml:"user_domain_id"`

	// ProjectDomainName is the name of the domain where a project resides.
	// It is used to identify the source domain of a project.
	// ProjectDomainName can be used in addition to a ProjectName when scoping
	// a user to a specific project.
	ProjectDomainName string `yaml:"project_domain_name"`

	// ProjectDomainID is the name of the domain where a project resides.
	// It is used to identify the source domain of a project.
	// ProjectDomainID can be used in addition to a ProjectName when scoping
	// a user to a specific project.
	ProjectDomainID string `yaml:"project_domain_id"`

	// DomainName is the name of a domain which can be used to identify the
	// source domain of either a user or a project.
	// If UserDomainName and ProjectDomainName are not specified, then DomainName
	// is used as a default choice.
	// It can also be used be used to specify a domain-only scope.
	DomainName string `yaml:"domain_name"`

	// DomainID is the unique ID of a domain which can be used to identify the
	// source domain of eitehr a user or a project.
	// If UserDomainID and ProjectDomainID are not specified, then DomainID is
	// used as a default choice.
	// It can also be used be used to specify a domain-only scope.
	DomainID string `yaml:"domain_id"`

	// DefaultDomain is the domain ID to fall back on if no other domain has
	// been specified and a domain is required for scope.
	DefaultDomain string `yaml:"default_domain"`
}
