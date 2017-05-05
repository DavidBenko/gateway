FROM mono
RUN nuget install Newtonsoft.Json
RUN cp /Newtonsoft.Json.9.0.1/lib/net45/Newtonsoft.Json.dll .
ADD /main.cs main.cs
RUN mcs main.cs -r:/Newtonsoft.Json.9.0.1/lib/net45/Newtonsoft.Json.dll
ENTRYPOINT ["mono", "main.exe"]
