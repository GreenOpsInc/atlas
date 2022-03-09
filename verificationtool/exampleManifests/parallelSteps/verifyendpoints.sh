#!/bin/sh
# apk --no-cache add curl
sleep 10
for i in 1 2 3 4 5
do
   echo $SERVICE_INTERNAL_URL
   resp=$(curl $SERVICE_INTERNAL_URL:5000/)
   if [ "$resp" != "Hello, World!" ]; then
      echo "Did not return correctly."
      exit 1;
   fi
   sleep 1
done
echo "Test completed correctly."
