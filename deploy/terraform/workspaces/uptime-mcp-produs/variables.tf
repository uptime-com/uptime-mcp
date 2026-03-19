variable "aws_region" {
  type = string
}

variable "eks_cluster_name" {
  type = string
}

variable "ecr_repositories" {
  type = list(object({
    name       = string
    mutability = optional(string, "IMMUTABLE")
  }))
}

variable "ref" {
  type = object({
    digest       = optional(string)
    semver       = optional(string)
    semverFilter = optional(string)
    tag          = optional(string)
  })
  validation {
    condition     = length(compact([var.ref.digest, var.ref.semver, var.ref.tag])) == 1
    error_message = "Exactly one of 'digest', 'semver', or 'tag' must be specified in oci_repository_ref."
  }
}
