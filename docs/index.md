# Astra Bridge

Astra Bridge is a system that allows for the transfer of ERC20 tokens between the Ethereum and MultiversX networks. The system is composed of several contracts and relayers that work together to facilitate the transfer of tokens.

## Ethereum Contracts
- **Repo**: https://github.com/multiversx/mx-bridge-eth-sc-sol
- **Safe (1)**: A contract that allows users to deposit ERC20 tokens that they want to transfer to the MultiversX network.
- **Bridge(2)**: A contract that facilitates the transfer of tokens from Ethereum to MultiversX.

## MultiversX Contracts
- **Repo**: https://github.com/multiversx/mx-bridge-eth-sc-rs
- **Safe (3)**: A contract that allows users to deposit ESDT tokens that they want to transfer back to the Ethereum network.
- **Bridge (4)**: A contract that facilitates the transfer of tokens from MultiversX to Ethereum.
- **MultiTransfer (5)**: A helper contract that is used to perform multiple token transfers at once.
- **BridgedTokensWrapper (6)**: A helper contract that is used to support wrap the same token from multiple chains into a single ESDT token.

## Relayers
- **Repo**: https://github.com/multiversx/mx-bridge-eth-go
- **5 Relayers**: Managed by the MultiversX Foundation.
- **5 Relayers**: Distributed to the MultiversX validators community, with each validator having one relayer.

## Transfer Flow

### Ethereum to MultiversX
1. A user deposits the ERC20 tokens that they want to transfer to the MultiversX network on the **Safe(1)** contract.
2. The **Safe(1)** contract groups multiple deposits into batches.
3. After a certain period of time, each batch becomes final and is processed by the relayers.
4. The relayers propose, vote, and perform the transfer using the **Bridge (4)** contract with a consensus of 7/10 votes.
5. The user receives the equivalent amount of ESDT tokens on their recipient address on the MultiversX network.
6. On the MultiversX network, the same amount of ESDT tokens are minted as were deposited on the Ethereum network.

### MultiversX to Ethereum
1. A user deposits the ESDT tokens that they want to transfer back to the Ethereum network on the **Safe(3)** contract.
2. The **Safe(3)** contract groups multiple deposits into batches.
3. After a certain period of time, each batch becomes final and is processed by the relayers.
4. The relayers propose, vote, and perform the transfer using the **Bridge (2)** contract with a consensus of 7/10 votes.
5. The user receives the equivalent amount of ERC20 tokens on their recipient address on the Ethereum network.
6. On the MultiversX network, the ESDT tokens that were transferred are burned.

## Support for Multiple Chains for the Same Token
The **BridgedTokensWrapper (6)** contract facilitates the use case of having the same token on multiple chains. It accepts the chain-specific ESDT token and mints a universal ESDT token that can be used on any application within the MultiversX network. The universal ESDT token can be converted back to the chain-specific ESDT token using the **BridgedTokensWrapper (6)** contract. This process burns the given universal tokens and sends the chain-specific ESDT tokens to the user.

Internally, the Astra Bridge system uses the **BridgedTokensWrapper (6)** contract to wrap the chain-specific tokens minted by the **MultiTransfer (5)** contract from multiple chains into a single ESDT token and sends it to the user.

When a user wants to transfer the tokens back to the source network, they must send the universal ESDT token to the **BridgedTokensWrapper (6)** contract, and the chain-specific ESDT token will be sent to the user. After this step, the user can send the chain-specific ESDT token to the **Safe (3)** contract, and the transfer will be performed as described in the previous section.

## Token Bridging Requirements
1. The MultiversX team must whitelist the token on both the Safe(1) and Safe(3) contracts. Only whitelisted tokens can be bridged.
2. The token issuer must issue the token on the MultiversX network and submit a branding request manually or using https://assets.multiversx.com/.
3. The token issuer must assign the MINT&BURN role to the BridgedTokensWrapper (6) contract as per the instructions provided at https://docs.multiversx.com/tokens/esdt-tokens/#setting-and-unsetting-special-roles

**Note**: As an alternative approach, MultiversX team can issue an ESDT token on the MultiversX chain with the same properties as on Ethereum, and give the needed roles to the Smart Contracts, as indicated above. The MultiversX team can then give the token issuer the ownership of token management for that specific ESDT token.

