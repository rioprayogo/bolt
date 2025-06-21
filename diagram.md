graph TD
    aws_local_vpc-main{{vpc-main}}
    aws_local_subnet-public[subnet-public]
    aws_local_web-sg(web-sg)
    aws_local_web-vm>web-vm]
    aws_local_vpc-main --> aws_local_subnet-public
    aws_local_vpc-main --> aws_local_web-sg
    aws_local_subnet-public --> aws_local_web-vm
    aws_local_web-sg --> aws_local_web-vm
