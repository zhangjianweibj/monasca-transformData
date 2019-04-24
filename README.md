=======================================
Docker image for Monasca data transform
=======================================
The transform-data image is based on the golang:1.12-alpine3.9.

Build arguments (child)
~~~~~~~~~~~~~~~~~~~~~~~
====================== =========================
Arguments              Occurrence
====================== =========================
BASE_TAG               Dockerfile
====================== =========================

Environment variables (child)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
============================== =============================== ================================================
Variable                       Default                         Description
============================== =============================== ================================================
LOG_LEVEL                      INFO                            logging level
CONSUMER_TOPIC                 custom-metrics                  get metrics from that topic
PRODUCER_TOPIC                 metrics                         send messages to the topic
KAFKA_URI                      kafka:9092                      The host and port for kafka
KAFKA_GROUP_ID 	               monasca-transform-data          the group id of consumer
ADMIN_ID                       admin                           id which replace metrics tenant_id property
POOL_FACTOR                    100                             size of metrics cached pool
MAX_THREADS                    1                               number of producer,which send messages concurrently
============================== =============================== ================================================

install in kubernetes environment (child)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
kubectl create -f monasca-transform-data-deployment.yml
