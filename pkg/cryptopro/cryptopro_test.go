package cryptopro

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/volodya-nrg/tools/pkg/exec_command"
)

func TestCryptoPro(t *testing.T) {
	t.Parallel()

	const (
		data = `{
			"data1": "123",
			"data2": "abc"
		}`
		cryptoCpFilepath = "/opt/cprocsp/bin/cryptcp"
	)

	crp := NewCryptoPro(exec_command.NewExecCommand(), cryptoCpFilepath)

	// зашифруем
	// личный контейнер (серт, подпись) должен уже присудствовать
	encrData, err := crp.Encrypt(t.Context(), []byte(data))
	require.NoError(t, err)
	require.NotEmpty(t, encrData)

	// расшифруем
	decrData, err := crp.Decrypt(t.Context(), encrData)
	require.NoError(t, err)
	require.NotEmpty(t, decrData)

	// проверим итоговый результат
	require.JSONEq(t, data, string(decrData))
}
