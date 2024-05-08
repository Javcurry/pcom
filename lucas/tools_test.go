package lucas_test

import (
	"hago-plat/pcom/lucas"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUs(t *testing.T) {
	assert.Equal(t, "api_version", lucas.CamelCaseToUnderscore("APIVersion"))
	assert.Equal(t, "yyid", lucas.CamelCaseToUnderscore("YYID"))
	assert.Equal(t, "yy_id", lucas.CamelCaseToUnderscore("YyID"))
	assert.Equal(t, "api_yy_id", lucas.CamelCaseToUnderscore("APIYyID"))
	assert.Equal(t, "apiyyid", lucas.CamelCaseToUnderscore("APIYYID"))
	assert.Equal(t, "d", lucas.CamelCaseToUnderscore("D"))
	assert.Equal(t, "version_d", lucas.CamelCaseToUnderscore("VersionD"))
	assert.Equal(t, "ap33_i_version", lucas.CamelCaseToUnderscore("AP33IVersion"))
	assert.Equal(t, "ap33i_version", lucas.CamelCaseToUnderscore("AP33iVersion"))
	assert.Equal(t, "ap33_i_version", lucas.CamelCaseToUnderscore("Ap33IVersion"))
	assert.Equal(t, "ap33_ai_version", lucas.CamelCaseToUnderscore("Ap33AIVersion"))
	assert.Equal(t, "api_yy_id", lucas.CamelCaseToUnderscore("api_yy_id"))
	assert.Equal(t, "_yy_id", lucas.CamelCaseToUnderscore("_yy_id"))
	assert.Equal(t, "api_yy.id", lucas.CamelCaseToUnderscore("api_yy.id"))
	assert.Equal(t, "api_yy.id", lucas.CamelCaseToUnderscore("api_yy.Id"))
	assert.Equal(t, "api_yy id", lucas.CamelCaseToUnderscore("api_yy Id"))
	assert.Equal(t, "api_yy_id", lucas.CamelCaseToUnderscore("api_yy_Id"))
	assert.Equal(t, "api_yy 你d", lucas.CamelCaseToUnderscore("api_yy 你d"))
}
