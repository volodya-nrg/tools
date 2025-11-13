package cryptopro

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type CryptoPro struct {
	executeCommander executeCommander
	cryptCpFilepath  string
}

func (c CryptoPro) Encrypt(ctx context.Context, data []byte) ([]byte, error) {
	fileIn, err := c.createFile(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("error creating file (in): %w", err)
	}

	defer func() {
		if err = os.Remove(fileIn.Name()); err != nil {
			slog.ErrorContext(ctx, "failed to remove temp-file (in)",
				slog.String("error", err.Error()),
				slog.String("temp-file", fileIn.Name()),
			)
		}
	}()

	fileOut, err := c.createFile(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating file (out): %w", err)
	}

	defer func() {
		if err = os.Remove(fileOut.Name()); err != nil {
			slog.ErrorContext(ctx, "failed to remove temp-file (out)",
				slog.String("error", err.Error()),
				slog.String("temp-file", fileIn.Name()),
			)
		}
	}()

	// -encr - кодируем
	// -errchain - если в цепочке ошибка, то выдать ошибку
	// -der - результат в виде байтов, для не больших данных (до 1000 символов)
	cmd := fmt.Sprintf("%s -encr -errchain -der %s %s", c.cryptCpFilepath, fileIn.Name(), fileOut.Name())

	if err = c.executeCommander.CommandRun(ctx, cmd); err != nil {
		return nil, fmt.Errorf("error executing command: %w", err)
	}

	result, err := os.ReadFile(fileOut.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to reading temp-file (out): %w", err)
	}

	return result, nil
}

func (c CryptoPro) Decrypt(ctx context.Context, data []byte) ([]byte, error) {
	fileIn, err := c.createFile(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("error creating file (in): %w", err)
	}

	defer func() {
		if err = os.Remove(fileIn.Name()); err != nil {
			slog.ErrorContext(ctx, "failed to remove temp-file (in)",
				slog.String("error", err.Error()),
				slog.String("temp-file", fileIn.Name()),
			)
		}
	}()

	fileOut, err := c.createFile(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating file (out): %w", err)
	}

	defer func() {
		if err = os.Remove(fileOut.Name()); err != nil {
			slog.ErrorContext(ctx, "failed to remove temp-file (out)",
				slog.String("error", err.Error()),
				slog.String("temp-file", fileIn.Name()),
			)
		}
	}()

	// -decr - расшифровываем
	cmd := fmt.Sprintf("%s -decr %s %s", c.cryptCpFilepath, fileIn.Name(), fileOut.Name())

	if err = c.executeCommander.CommandRun(ctx, cmd); err != nil {
		return nil, fmt.Errorf("error executing command: %w", err)
	}

	result, err := os.ReadFile(fileOut.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to reading temp-file (out): %w", err)
	}

	return result, nil
}

func (c CryptoPro) createFile(ctx context.Context, data []byte) (*os.File, error) {
	f, err := os.CreateTemp("", "cryptopro-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp-file: %w", err)
	}

	defer func() {
		if err = f.Close(); err != nil {
			slog.ErrorContext(ctx, "failed to close temp-file")
		}
	}()

	if len(data) > 0 {
		if _, err = f.Write(data); err != nil {
			return nil, fmt.Errorf("failed to write data to temp-file (%s): %w", f.Name(), err)
		}
	}

	return f, nil
}

func NewCryptoPro(executeCommander executeCommander, cryptCpFilepath string) *CryptoPro {
	return &CryptoPro{
		executeCommander: executeCommander,
		cryptCpFilepath:  cryptCpFilepath,
	}
}
