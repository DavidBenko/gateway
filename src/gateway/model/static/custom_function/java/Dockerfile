FROM openjdk
ADD /Main.java Main.java
RUN wget -O json.jar https://search.maven.org/remotecontent?filepath=org/json/json/20160810/json-20160810.jar
RUN javac -cp json.jar Main.java
ENTRYPOINT ["java", "-cp", "json.jar:.", "Main"]
