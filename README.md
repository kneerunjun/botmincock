# botmincock
telegram bot backend for accounts maintenance, patching the payment gateway, monthly book keeping


docker-compose --env-file dev.env up mongostore 

docker-compose --env-file dev.env up botmincock 
docker-compose --env-file dev.env build

docker exec -it ctn_store /bin/bash -c mongo


dev-btmnck
b07m1nc0ck