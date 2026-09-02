# Import an account-level tracking domain
terraform import sparkpost_tracking_domain.example track.example.com

# Import a tracking domain that belongs to a subaccount
terraform import sparkpost_tracking_domain.example track.example.com,12345
