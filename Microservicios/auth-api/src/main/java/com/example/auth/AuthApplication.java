package com.example.auth;

import jakarta.annotation.PostConstruct;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.io.File;
import java.sql.*;
import java.time.Instant;
import java.util.Map;

@SpringBootApplication
@RestController
public class AuthApplication {

    private static final String DB_PATH = System.getenv().getOrDefault("DB_PATH", "/data/auth.db");
    private static final String JDBC_URL = "jdbc:sqlite:" + DB_PATH;

    public static void main(String[] args) {
        SpringApplication.run(AuthApplication.class, args);
    }

    @PostConstruct
    public void initDb() throws Exception {
        File dbFile = new File(DB_PATH);
        if (dbFile.getParentFile() != null) {
            dbFile.getParentFile().mkdirs();
        }
        try (Connection conn = DriverManager.getConnection(JDBC_URL);
             Statement st = conn.createStatement()) {
            st.execute("""
                CREATE TABLE IF NOT EXISTS credentials (
                    email TEXT PRIMARY KEY,
                    password TEXT NOT NULL
                )
                """);
            st.execute("""
                CREATE TABLE IF NOT EXISTS auth_logs (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    email TEXT NOT NULL,
                    valid INTEGER NOT NULL,
                    timestamp TEXT NOT NULL
                )
                """);
            st.execute("""
                INSERT OR IGNORE INTO credentials (email, password)
                VALUES ('admin@test.com', '1234')
                """);
        }
    }

    @GetMapping("/health")
    public Map<String, Object> health() {
        return Map.of("service", "auth-api", "status", "UP");
    }

    @PostMapping("/auth/login")
    public ResponseEntity<Map<String, Object>> login(@RequestBody Map<String, String> body) throws Exception {
        String email = body.getOrDefault("email", "");
        String password = body.getOrDefault("password", "");
        boolean isValid = false;

        try (Connection conn = DriverManager.getConnection(JDBC_URL);
             PreparedStatement ps = conn.prepareStatement(
                     "SELECT 1 FROM credentials WHERE email = ? AND password = ?")) {
            ps.setString(1, email);
            ps.setString(2, password);
            try (ResultSet rs = ps.executeQuery()) {
                isValid = rs.next();
            }
        }

        try (Connection conn = DriverManager.getConnection(JDBC_URL);
             PreparedStatement logPs = conn.prepareStatement(
                     "INSERT INTO auth_logs (email, valid, timestamp) VALUES (?, ?, ?)")) {
            logPs.setString(1, email);
            logPs.setInt(2, isValid ? 1 : 0);
            logPs.setString(3, Instant.now().toString());
            logPs.executeUpdate();
        }

        if (isValid) {
            return ResponseEntity.ok(Map.of(
                    "valid", true,
                    "message", "Credenciales validas",
                    "email", email
            ));
        }
        return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of(
                "valid", false,
                "message", "Credenciales invalidas"
        ));
    }
}