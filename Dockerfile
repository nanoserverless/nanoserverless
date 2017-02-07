FROM    	scratch
COPY		nanoserverless /
ENTRYPOINT	["/nanoserverless"]
EXPOSE 		80
