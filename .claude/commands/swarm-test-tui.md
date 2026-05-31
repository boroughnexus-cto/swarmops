Run the TUI-specific test suite with verbose output.

```bash
cd /home/sbarker/git-bnx/TKN/swarmops && go test -v -run 'TestView|TestKey|TestSidebar|TestWindow|TestTopBar|TestStatus|TestContent|TestMouse|TestClassify|TestAnimated|TestIntegration|TestHandleKey|TestLoadItems|TestSmoke' -timeout 60s ./...
```

Report:
- Pass/fail counts
- Any failing test names + first error line
- Any panics or race conditions
- Total elapsed time

To regenerate golden files after intentional rendering changes:
```bash
UPDATE_GOLDEN=1 go test -run 'TestView' ./...
```
