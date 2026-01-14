# Terraform Provider Rustfs

This provider should take care of rustfs management.
The issue I needed to solve is a missing functionality of the aws provider on my installation.
A working solution was: https://github.com/aminueza/terraform-provider-minio for bucket management.

As IAM should also be maneged I will use the provider of aminueza and add some rust endpoint.



## Endpoints
At the moment I was not able to get a good api definition. Will try: https://deepwiki.com/rustfs/rustfs/10-reference#admin-api-routes and some postman magic.


## Acceptanc testing

Unit test performed with pkg work simply out of the box.
To perform acceptance test of the provider we need to follow: https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install

 ~/.terraformrc
```
provider_installation {
  dev_overrides {
      "weinmann/rustfs" = "/workspaces/terraform-provider-rustfs"
  }
}
```

## Publish
ToDo: https://developer.hashicorp.com/terraform/registry/providers/publishing
