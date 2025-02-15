---
layout: "yandex"
page_title: "Yandex: yandex_kms_asymmetric_encryption_key"
sidebar_current: "docs-yandex-kms-asymmetric-encryption-key"
description: |-
Creates a Yandex KMS asymmetric encryption key that can be used for cryptographic operation.
---

# yandex\_kms\_asymmetric\_encryption\_key

Creates a Yandex KMS asymmetric encryption key that can be used for cryptographic operation.

~> **Note:** When Terraform destroys this key,
any data previously encrypted with this key will be irrecoverable.
For this reason, it is strongly recommended that you add lifecycle hooks
to the resource to prevent accidental destruction.

For more information, see [the official documentation](https://cloud.yandex.com/docs/kms/concepts/).

## Example Usage

```hcl
resource "yandex_kms_asymmetric_encryption_key" "key-a" {
  name              = "example-asymetric-encryption-key"
  description       = "description for key"
  encryption_algorithm = "RSA_2048_ENC_OAEP_SHA_256"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the key.

* `description` - (Optional) An optional description of the key.

* `folder_id` - (Optional) The ID of the folder that the resource belongs to. If it
  is not provided, the default provider folder is used.

* `labels` - (Optional) A set of key/value label pairs to assign to the key.

* `encryption_algorithm` - (Optional) Encryption algorithm to be used with a new key. The default value is `RSA_2048_ENC_OAEP_SHA_256`.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `status` - The status of the key.
* `created_at` - Creation timestamp of the key.

## Timeouts

`yandex_kms_asymmetric_encryption_key` provides the following configuration options for
[timeouts](/docs/configuration/resources.html#timeouts):

- `create` - Default 1 minute
- `update` - Default 1 minute
- `delete` - Default 1 minute

## Import

A KMS asymmetric encryption key can be imported using the `id` of the resource, e.g.

```
$ terraform import yandex_kms_asymmetric_encryption_key.top-secret kms_asymmetric_encryption_key_id
```

