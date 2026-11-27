FROM maven:3.9-eclipse-temurin-21 AS build
WORKDIR /app
COPY pom.xml .
COPY src ./src
RUN mvn -q package -DskipTests

FROM eclipse-temurin:21-jre
WORKDIR /app
COPY --from=build /app/target/auth-api-1.0.0.jar app.jar
EXPOSE 8001

CMD ["java", "-Dserver.port=8001", "-jar", "app.jar"]