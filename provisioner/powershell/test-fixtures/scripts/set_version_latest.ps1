Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function test
{

    $myNumbersCollection = 1..5

    if($myNumbersCollection -contains 3)
    {
        "collection contains 3"
    }
    else
    {
        "collection doesn't contain 3"
    }
}

test
