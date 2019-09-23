<!-- Code generated from the comments of the TargetImageConfig struct in builder/triton/target_image_config.go; DO NOT EDIT MANUALLY -->

-   `image_description` (string) - Description of the image. Maximum 512
    characters.
    
-   `image_homepage` (string) - URL of the homepage where users can find
    information about the image. Maximum 128 characters.
    
-   `image_eula_url` (string) - URL of the End User License Agreement (EULA)
    for the image. Maximum 128 characters.
    
-   `image_acls` ([]string) - The UUID's of the users which will have
    access to this image. When omitted only the owner (the Triton user whose
    credentials are used) will have access to the image.
    
-   `image_tags` (map[string]string) - Tag applied to the image.
    