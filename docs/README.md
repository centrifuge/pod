# Centrifuge POD (Private Off-chain Data) Description

The purpose of this document is to provide more information regarding the basic functionality of the POD.

# Table of Contents
1. [Accounts](#accounts)
2. [Documents](#documents)
3. [NFTs](#nfts)
4. [Jobs](#jobs)
5. [Webhooks](#webhooks)
6. [Identities](#identities)

## Accounts

[//]: # (TODO&#40;cdamian&#41;: Update the swagger link to latest version.)
The `Accounts` section of our [swagger API docs](https://app.swaggerhub.com/apis/centrifuge.io/cent-node/2.1.0#/Accounts) provides
an overview of all the endpoints available for handling accounts.

---

An account is the POD representation of the user that is performing various operations. This identity of this account
is used when storing documents and performing any action related to the document handling process such as - starting long-running
tasks for committing or minting documents, or sending the document via the p2p layer.

### Data

The data stored for each account has the following JSON format:

```json
{
  "data": [
    {
      "identity": "string",
      "document_signing_public_key": [
        0
      ],
      "p2p_public_signing_key": [
        0
      ],
      "pod_operator_account_id": [
        0
      ],
      "precommit_enabled": true,
      "webhook_url": "string"
    }
  ]
}
```

`identity` - hex encoded Centrifuge Chain account ID. This is the identity used for performing the operation described above.

`document_signing_public_key` - public key that is used for signing documents.

`p2p_public_signing_key` - public key that is used for interactions on the p2p layer.

`pod_operator_account_id` - the POD operator account ID. See [pod operator](#pod-operator).

`precommit_enabled` - flag that enables anchoring the document prior to requesting the signatures from all collaborators.

`webhook_url` - URL of the webhook that is used for sending updates regarding documents or jobs.

### Account Creation

[//]: # (TODO&#40;cdamian&#41;: Update the swagger link to latest version.)

An account can be created by calling the [endpoint](https://app.swaggerhub.com/apis/centrifuge.io/cent-node/2.1.0#/Accounts/generate_account_v2) and providing
the required information - `identity`, `precommit_enabled`, `webhook_url`.

The successful response for the account creation operation will contain the extra fields mentioned above in [Data](#data).

### Account Boostrap

**IMPORTANT** - The following steps are required to ensure that the POD can use a newly created identity.

1. Store the `document_signing_public_key` and `p2p_public_signing_key` in the `Keystore` of Centrifuge Chain.

   This can be done by submitting the `addKeys` extrinsic of the `Keystore` pallet.
2. Add the POD operator account ID as a `PodOperation` proxy to the `identity`.

   This can be done by submitting the `addProxy` extrinsic of the `Proxy` pallet.

## Documents

[//]: # (TODO&#40;cdamian&#41;: Update the swagger link to latest version.)
The `Documents` section of our [swagger API docs](https://app.swaggerhub.com/apis/centrifuge.io/cent-node/2.1.0#/Documents) provides
an overview of all the endpoints available for handling documents.

---

The main purpose of the POD is to serve as a handler for documents that contain private off-chain data.

## NFTs

[//]: # (TODO&#40;cdamian&#41;: Update the swagger link to latest version.)
The `NFTs` section of our [swagger API docs](https://app.swaggerhub.com/apis/centrifuge.io/cent-node/2.1.0#/NFTs) provides
an overview of all the endpoints available for handling document NFTs.

---

The NFT endpoint provide basic functionality for minting NFTs for a document and retrieving NFT specific information
such as attributes, metadata, and owner.

## Jobs

[//]: # (TODO&#40;cdamian&#41;: Update the swagger link to latest version.)
The `Jobs` section of our [swagger API docs](https://app.swaggerhub.com/apis/centrifuge.io/cent-node/2.1.0#/Jobs) provides
an overview of all the endpoints available for retrieving job details.

---

The jobs endpoint returns detailed information for a job. 

A job is a long-running operation that is triggered by the POD when performing actions related to documents or NFTs.

## Webhooks

[//]: # (TODO&#40;cdamian&#41;: Update the swagger link to latest version.)
The `Webhook` section of our [swagger API docs](https://app.swaggerhub.com/apis/centrifuge.io/cent-node/2.1.0#/Webhook) provides
an overview the notification message that is sent by the POD for document or job events.

## Identities

Most of the operations performed by the POD rely on the presence of proxies that are used to:
- sign JSON Web3 Tokens used for [authentication](Authentication.md).
- sign extrinsics that are performed on behalf of the identity (see [account bootstrap](#boostrap)).

### POD-specific Accounts

#### POD Admin

The POD admin is an account that is stored on the POD, and its sole purpose is to authorize access for some account related endpoints such as
account generation, accounts listing, account details retrieval. 
This is required since not every user should have the rights to perform the mentioned actions.

#### POD Operator

The POD operator is an account that is stored on the POD, and it is used for submitting extrinsics on behalf of the provided identity. 
This is required since an identity can be an anonymous proxy, which is unable to sign any extrinsics.
