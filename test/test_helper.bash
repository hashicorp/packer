# Let's verify that the tools we need are installed
declare -a required=(aws)
for cmd in "${required[@]}"; do
    command -v $cmd >/dev/null 2>&1 || {
        echo "'$cmd' must be installed" >&2
        exit 1
    }
done

#--------------------------------------------------------------------
# Bats modification
#--------------------------------------------------------------------
# This allows us to override a function in Bash
save_function() {
    local ORIG_FUNC=$(declare -f $1)
    local NEWNAME_FUNC="$2${ORIG_FUNC#$1}"
    eval "$NEWNAME_FUNC"
}

# Override the run function so that we always output the output
save_function run old_run
run() {
    old_run $@

    # Output the command we ran
    echo "Executing: " $@

    # "$output" gets rid of newlines. This will bring them back.
    for line in "${lines[@]}"; do
        echo $line
    done
}

#--------------------------------------------------------------------
# Helper functions
#--------------------------------------------------------------------
# This sets the directory for fixtures by specifying the name of
# the folder with fixtures.
fixtures() {
    FIXTURE_ROOT="$BATS_TEST_DIRNAME/fixtures/$1"
}

# This deletes any AMIs with a tag "packer-test" of "true"
aws_ami_cleanup() {
    local region=${1:-us-east-1}
    aws ec2 describe-images --region ${region} --owners self --output text \
        --filters 'Name=tag:packer-test,Values=true' \
        --query 'Images[*].ImageId' \
        | xargs -n1 aws ec2 deregister-image --region ${region} --image-id
}
