# Use the official InfluxDB image
FROM influxdb:2.7-alpine

# Environment variables for initial setup
ENV DOCKER_INFLUXDB_INIT_MODE=setup
ENV DOCKER_INFLUXDB_INIT_USERNAME=admin
ENV DOCKER_INFLUXDB_INIT_PASSWORD=adminpassword
ENV DOCKER_INFLUXDB_INIT_ORG=my-org
ENV DOCKER_INFLUXDB_INIT_BUCKET=llm_metrics
ENV DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=my-super-secret-admin-token

# Expose InfluxDB port
EXPOSE 8086

# Data persistence volume
VOLUME /var/lib/influxdb2 