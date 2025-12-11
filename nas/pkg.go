/*
Package nas provides support for efficiently working with
network attached storage (NAS).
*/
package nas

import (
	"github.com/konveyor/tackle2-hub/shared/nas"
)

var MkDir = nas.MkDir
var CpDir = nas.CpDir
var RmDir = nas.RmDir
var HasDir = nas.HasDir
var Exists = nas.Exists
