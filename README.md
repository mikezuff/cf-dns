Exploring the Cloudflare DNS API

Add API key/token in ./keys:

```
echo "export CLOUDFLARE_API_TOKEN='1234567890123456789012345678999999999000'" > keys
echo "export CLOUDFLARE_API_EMAIL='user@example.com'" >> keys
echo "export CLOUDFLARE_API_KEY='2345678345678456783456783456784567842'" >> keys
```

Running

```
go build && ./cf-dns | jq .
```
