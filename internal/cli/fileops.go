package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JianLoong/robogo/internal/actions"
	"github.com/JianLoong/robogo/internal/parser"
	"github.com/JianLoong/robogo/internal/runner"
)

// FileProcessor handles file discovery and processing
type FileProcessor struct {
	executor *actions.ActionExecutor
	options  RunOptions
}

// NewFileProcessor creates a new file processor
func NewFileProcessor(executor *actions.ActionExecutor, options RunOptions) *FileProcessor {
	return &FileProcessor{
		executor: executor,
		options:  options,
	}
}

// ProcessPaths processes multiple file paths and returns aggregated results
func (fp *FileProcessor) ProcessPaths(ctx context.Context, paths []string) (*RunResults, error) {
	results := &RunResults{}

	for _, path := range paths {
		pathResults, err := fp.processPath(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("failed to process path %s: %w", path, err)
		}

		results.SuiteResults = append(results.SuiteResults, pathResults.SuiteResults...)
		results.CaseResults = append(results.CaseResults, pathResults.CaseResults...)
	}

	return results, nil
}

// processPath processes a single path (file or directory)
func (fp *FileProcessor) processPath(ctx context.Context, path string) (*RunResults, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s: %w", path, err)
	}

	if info.IsDir() {
		return fp.processDirectory(ctx, path)
	}

	return fp.processFile(ctx, path)
}

// processDirectory recursively processes all .robogo files in a directory
func (fp *FileProcessor) processDirectory(ctx context.Context, dirPath string) (*RunResults, error) {
	results := &RunResults{}

	err := filepath.Walk(dirPath, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fileInfo.IsDir() || !strings.HasSuffix(filePath, ".robogo") {
			return nil
		}

		fileResults, err := fp.processFile(ctx, filePath)
		if err != nil {
			return err
		}

		results.SuiteResults = append(results.SuiteResults, fileResults.SuiteResults...)
		results.CaseResults = append(results.CaseResults, fileResults.CaseResults...)

		return nil
	})

	return results, err
}

// processFile processes a single .robogo file
func (fp *FileProcessor) processFile(ctx context.Context, filePath string) (*RunResults, error) {
	isSuite, err := fp.isTestSuite(filePath)
	if err != nil {
		return nil, err
	}

	if isSuite {
		return fp.processSuiteFile(ctx, filePath)
	}

	return fp.processCaseFile(ctx, filePath)
}

// processSuiteFile processes a test suite file
func (fp *FileProcessor) processSuiteFile(ctx context.Context, filePath string) (*RunResults, error) {
	testSuite, err := parser.ParseTestSuite(filePath)
	if err != nil {
		return nil, err
	}

	testExecutor := runner.NewTestExecutionService(fp.executor)
	if fp.options.VariableDebug {
		testExecutor.GetContext().EnableVariableDebugging(true)
	}

	suiteRunner := runner.NewTestSuiteRunner(testExecutor)
	result, err := suiteRunner.RunTestSuite(ctx, testSuite, filePath, fp.options.Silent)
	if err != nil {
		return nil, err
	}

	return &RunResults{
		SuiteResults: []*parser.TestSuiteResult{result},
	}, nil
}

// processCaseFile processes a test case file
func (fp *FileProcessor) processCaseFile(ctx context.Context, filePath string) (*RunResults, error) {
	results, err := runner.RunTestFilesWithConfigAndDebug(
		ctx,
		[]string{filePath},
		fp.options.Silent,
		fp.options.ParallelConfig,
		fp.executor,
		fp.options.VariableDebug,
	)
	if err != nil {
		return nil, err
	}

	return &RunResults{
		CaseResults: results,
	}, nil
}

// isTestSuite determines if a file is a test suite by examining its content
func (fp *FileProcessor) isTestSuite(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buffer := make([]byte, 4096)
	n, _ := file.Read(buffer)
	content := string(buffer[:n])

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "testsuite:") {
			return true, nil
		}

		if strings.HasPrefix(line, "testcase:") {
			return false, nil
		}

		// Fallback: check for testcases array
		if strings.HasPrefix(line, "testcases:") {
			return true, nil
		}

		break
	}

	return false, nil
}
