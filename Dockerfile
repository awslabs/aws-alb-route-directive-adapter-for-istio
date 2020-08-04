FROM scratch
ADD authzadaptor /authzadaptor
EXPOSE 9070
ENTRYPOINT ["/authzadaptor"]
