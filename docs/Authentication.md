# Centrifuge POD Authentication

Authentication is performed using the JSON Web3 Tokens described [here](https://github.com/hamidra/jw3t).

## Token

### Format

The format of the JW3 token that we use is:

`base_64_encoded_json_header.base_64_encoded_json_payload.base_64_encoded_signature`

Where the un-encoded parts are as follows:

Header:

```json
{
  "algorithm": "sr25519", 
  "token_type": "JW3T", 
  "address_type": "ss58"
}
```

---

Payload:

```json
{
  "address": "delegate_address",
  "on_behalf_of": "delegator_address",
  "proxy_type": "proxy_type",
  "expires_at": "1663070957",
  "issued_at": "1662984557",
  "not_before": "1662984557"
}
```

`address` - SS58 address of the proxy delegate (see [usage](#usage) for more info).

`on_behalf_of` - SS58 address of the proxy delegator (see [usage](#usage) for more info).

`proxy_type` - one of the allowed proxy types (see [usage](#usage) for more info):

- `PodAdmin` - defined in the POD.
- `Any` - defined in the Centrifuge Chain.
- `PodOperation` - defined in the Centrifuge Chain.
- `PodAuth` - defined in the Centrifuge Chain.

`expires_at` - token expiration time.

`issued_at` - token creation time.

`not_before` - token activation time.

---

Signature - the `Schnorrkel/Ristretto x25519` signature generated for `json_header.json_payload`.

---

**NOTE** - An example on how to generate a JW3 token can be found [here](../http/auth/test_utils.go).

### Usage

The POD has 2 types of authentication mechanisms:

1. On-chain proxies - this is the most commonly used mechanism and it's used to authenticate any on-chain proxies of the identity. 
  
   In this case, the `address`, `on_behalf_of` and `proxy_type` should contain the information as found on-chain.

   Example:

   `Alice` - identity.
   
   `Bob` - proxy of `Alice` with type `Any`.

   Token payload:

   ```json
   {
     "address": "ss58_address_of_bob",
     "on_behalf_of": "ss58_address_of_alice",
     "proxy_type": "Any",
     "expires_at": "1663070957",
     "issued_at": "1662984557",
     "not_before": "1662984557"
   }
   ```

2. POD admin - this is used when performing authentication for restricted endpoints.

   In this case, the `address` and `on_behalf_of` fields should be equal and contain the SS58 address of the POD admin, and
   the `proxy_type` should be `PodAdmin`.

   Example:

   ```json
   {
     "address": "pod_admin_ss58_address",
     "on_behalf_of": "pod_admin_ss58_address",
     "proxy_type": "PodAdmin",
     "expires_at": "1663070957",
     "issued_at": "1662984557",
     "not_before": "1662984557"
   }
   ```