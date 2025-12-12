package identity

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests and reuse.
var (
	GitPw = api.Identity{
		Kind:     "git",
		Name:     "git-pw",
		User:     "test-user",
		Password: "test-password-123",
	}
	GitKey = api.Identity{
		Kind: "git",
		Name: "git-key",
		User: "test-user",
		Key: /* The key was generated only for test purposes, not used anywhere, gitleaks:allow */ `-----BEGIN OPENSSH PRIVATE KEY-----
		b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAlwAAAAdzc2gtcn
		NhAAAAAwEAAQAAAIEAoDrNL4XSvv5xgPX4ty4nZFn+5nTG9F9Kvx8u+eNN2CdTKjzFUr0N
		HYFfiBo98RyZO67plzQVi7hK+ooUTAdt+eU6bgF23VF62Cd8GT/l2AMB73tTedEYVmsOi5
		aJwJV3BaLWmgEbg3YqbtJWeYHloWV1myC7ChYFIqgNgq/NikcAAAIAQpLNF0KSzRcAAAAH
		c3NoLXJzYQAAAIEAoDrNL4XSvv5xgPX4ty4nZFn+5nTG9F9Kvx8u+eNN2CdTKjzFUr0NHY
		FfiBo98RyZO67plzQVi7hK+ooUTAdt+eU6bgF23VF62Cd8GT/l2AMB73tTedEYVmsOi5aJ
		wJV3BaLWmgEbg3YqbtJWeYHloWV1myC7ChYFIqgNgq/NikcAAAADAQABAAAAgC9qd0Hy9j
		SzE1+4aou6ysl1oGsOb7ntYS9BySMF0wl+SawcI8CiVXk/ycd7DI2HiwkQb+vg6+dW+ElA
		B1WgdTigUZjmhJ8nAFgebav8PbTDc2izFGu31jLeP6dyLRcRG/ztcOvsglloh9Oj//jKzb
		YBpKlhIaJbnmVokx0NgAMxAAAAQQCc8TsykYsJQz8R2MpddptTlIStFRIvt7bVcwZ8Hrr5
		M8XT3aCxABvtcRgY++/VDZBkK+TNfcdWSS+1Cxrmp/X2AAAAQQDT109BFaf9bbmOdipdyd
		nXFhQgxwZ2HbCBtMmOR2zEosCPDNf2c7McFTAOCM3Mh0qNlZPPbUL1v2TVWvDi9NQjAAAA
		QQDBoU1fSYFq+Rf2Dyrq9RevXT4VO+uahlje46zV2sVA6Mg2lc+yqTbTIv9NvxnJMC51vO
		B1q2YQPNuJWgd03jGNAAAACm1hdWZhcnRAbmI=
		-----END OPENSSH PRIVATE KEY-----
		`,
	}
	Mvn = api.Identity{
		Kind: "maven",
		Name: "maven-settings",
		Settings: `<?xml version="1.0" encoding="UTF-8"?>
		<settings xmlns="http://maven.apache.org/SETTINGS/1.2.0"
				  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
				  xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.2.0 http://maven.apache.org/xsd/settings-1.2.0.xsd">
		  <pluginGroups>
		  </pluginGroups>
		  <proxies>
		  </proxies>
		  <servers>
			<server>
			   <id>tackle-testapp-public</id>
			   <username>GITHUB_USER</username>
			   <password>GITHUB_TOKEN</password>
			 </server>
		  </servers>
		  <mirrors>
			<mirror>
			  <id>maven-default-http-blocker</id>
			  <mirrorOf>external:http:*</mirrorOf>
			  <name>Pseudo repository to mirror external repositories initially using HTTP.</name>
			  <url>http://0.0.0.0/</url>
			  <blocked>true</blocked>
			</mirror>
		  </mirrors>
		  <profiles>
			<profile>
			  <id>github</id>
			  <repositories>
				<repository>
				  <id>central</id>
				  <url>https://repo1.maven.org/maven2</url>
				</repository>
				<repository>
				  <id>tackle-testapp-public</id>
				  <url>https://maven.pkg.github.com/konveyor/tackle-testapp-public</url>
				  <snapshots>
					<enabled>true</enabled>
				  </snapshots>
				</repository>
			  </repositories>
			</profile>
		  </profiles>
		  <activeProfiles>
			<activeProfile>github</activeProfile>
		  </activeProfiles>
		</settings>`,
	}
	Proxy = api.Identity{
		Kind:     "proxy",
		Name:     "Proxy Identity",
		User:     "test-user",
		Password: "test-password",
	}
	BasicAuth = api.Identity{
		Kind:        "basic-auth",
		Description: "Basic Auth (Jira)",
		Name:        "Basic Auth Identity",
		User:        "test-user",
		Password:    "test-password",
	}
	BearerToken = api.Identity{
		Kind: "bearer",
		Name: "Bearer Token Identity",
		Key:  "test-key",
	}
	Samples = []api.Identity{GitPw, GitKey, Mvn, Proxy, BasicAuth, BearerToken}
)
