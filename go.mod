module github.com/konveyor/tackle2-hub

go 1.16

require (
	github.com/Nerzal/gocloak/v10 v10.0.1
	github.com/gin-gonic/gin v1.7.4
	github.com/golang-jwt/jwt/v4 v4.3.0
	github.com/google/uuid v1.1.2
	github.com/konveyor/controller v0.8.0
	github.com/mattn/go-sqlite3 v1.14.9
	github.com/onsi/gomega v1.7.0
	github.com/prometheus/client_golang v1.1.0
	github.com/swaggo/swag v1.7.8
	golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e
	gorm.io/datatypes v1.0.5
	gorm.io/driver/postgres v1.2.3 // indirect
	gorm.io/driver/sqlite v1.2.4
	gorm.io/driver/sqlserver v1.2.1 // indirect
	gorm.io/gorm v1.22.5
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/apiserver v0.17.1
	k8s.io/client-go v0.17.4
	sigs.k8s.io/controller-runtime v0.1.11
)

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20181127025237-2b1284ed4c93

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20181213151034-8d9ed539ba31

replace k8s.io/api => k8s.io/api v0.0.0-20181213150558-05914d821849

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181213153335-0fe22c71c476
