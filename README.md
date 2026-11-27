Miniactividad: Arquitectura de Microservicios con APIs Multilenguaje, SQLite y Docker

 # 1. Construir y levantar los 4 servicios
docker compose up -d --build

# 2. Ver estado y logs
docker compose ps
docker compose logs -f

# 3. Probar endpoints básicos
curl -X POST http://localhost:8000/users -H "Content-Type: application/json" \
  -d '{"name":"Ana","email":"ana@test.com","password":"secret"}'
curl http://localhost:8000/users

curl http://localhost:8001/health
curl -X POST http://localhost:8001/auth/login -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"1234"}'

curl -X POST http://localhost:8002/orders -H "Content-Type: application/json" \
  -d '{"customer":"Ana","total":45.5,"status":"CREATED"}'
curl http://localhost:8002/orders

curl -X POST http://localhost:8003/products -H "Content-Type: application/json" \
  -d '{"name":"Teclado","price":29.99,"description":"USB","stock":10}'
curl http://localhost:8003/products

# 4. Recrear contenedores y comprobar persistencia (sin -v en down)
docker compose down
docker compose up -d
curl http://localhost:8000/users
