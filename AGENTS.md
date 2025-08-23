## Development commands

```
# In background start the api backend air build that watches modifications and rebuilds
powershell -Command "Start-Process -FilePath 'bash' -ArgumentList './make', 'dev'"

# In background start the frontend vite build what also watches modifications
powershell -Command "Start-Process -FilePath 'npm' -ArgumentList 'run', 'dev'"
```
