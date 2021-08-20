variable "properties" {
    type = map(object({
        cpcode = string
        origin = string
        hostname = string
    }))
}