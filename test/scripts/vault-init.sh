docker-compose -f docker-compose-vault-test.yaml up
VAULT_ADDR='http://127.0.0.1:8200/' vault operator init -key-shares=1 -key-threshold=1
VAULT_ADDR='http://127.0.0.1:8200/' vault operator unseal 
VAULT_ADDR='http://127.0.0.1:8200/' vault operator login
??
