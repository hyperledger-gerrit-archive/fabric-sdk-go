docker-compose --file=./test/fixtures/docker-compose-unit.yaml ps -q | xargs docker inspect -f '{{ .State.ExitCode }}' | \
while read code; do  
    echo Found error code $code
    if (test $code -ne 0 ) then    
       exit 1
    fi
done  