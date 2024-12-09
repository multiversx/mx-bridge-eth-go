#!/bin/bash

# Call the endpoint and store the response in a variable
response=$(curl -s "http://localhost:8085/simulator/initial-wallets")

# Use jq to extract the "bech32" address from the JSON response
bech32_address=$(echo "$response" | jq -r '.data.balanceWallets | to_entries[0].value.address.bech32')

# Print the address to verify it was extracted correctly
echo "Extracted Bech32 Address: $bech32_address"

# Get account information
account_info_response=$(curl -s --request GET \
  --url "http://localhost:8085/address/$bech32_address" \
  --header 'User-Agent: insomnia/10.0.0')

# Print the response for debugging
echo "Account info response: $account_info_response"

# Step 2: Extract the nonce using jq
nonce=$(echo "$account_info_response" | jq -r '.data.account.nonce')

# Check if nonce is not empty
if [ -n "$nonce" ]; then
    echo "Extracted nonce: $nonce"
else
    echo "Error: No nonce found in the account info response."
    exit 1
fi

tx_send_response=$(curl -s --request POST \
  --url http://localhost:8085/transaction/send \
  --header 'Content-Type: application/json' \
  --header 'User-Agent: insomnia/10.0.0' \
  --data '{
	"nonce": '$nonce',
	"value": "50000000000000000000",
	"sender": "'$bech32_address'",
	"receiver": "erd12js50s7ycclwpac4qpx7lty3prhpu8hy00thjgz9f67p33w7m94qmzttem",
	"gasLimit": 50000,
	"gasPrice": 1000000000,
	"chainId": "chain",
	"signature": "aa",
	"version": 1
}')

echo $tx_send_response

# Extract txHash using jq
tx_hash=$(echo "$tx_send_response" | jq -r '.data.txHash')

# Print the extracted txHash
echo "Extracted txHash: $tx_hash"

sleep 2

# Make the second call to generate blocks until the transaction is processed
generate_blocks_response=$(curl -s --request POST \
  --url "http://localhost:8085/simulator/generate-blocks-until-transaction-processed/$tx_hash" \
  --header 'Content-Type: application/json' \
  --header 'User-Agent: insomnia/10.0.0')

# Print the response from the second call
echo "Response from generate blocks: $generate_blocks_response"
