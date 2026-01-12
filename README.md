# Terraform Provider Rustfs

This provider should take care of rustfs management.
The issue I needed to solve is a missing functionality of the aws provider on my installation.
A working solution was: https://github.com/aminueza/terraform-provider-minio for bucket management.

As IAM should also be maneged I will use the provider of aminueza and add some rust endpoint.



## Endpoints
At the moment I was not able to get a good api definition. Will try: https://deepwiki.com/rustfs/rustfs/10-reference#admin-api-routes and some postman magic.


