# lab6
Cargar Partidos
<img width="468" alt="image" src="https://github.com/user-attachments/assets/c841ad1f-0901-48ae-9dfe-ed0df2ef6cd0" />

Crear nuevo partido
<img width="456" alt="image" src="https://github.com/user-attachments/assets/a460577e-1157-4d0e-afce-113b8e9935ae" />


Actualizar partido
<img width="400" alt="image" src="https://github.com/user-attachments/assets/78972156-80ce-4f2b-a7a8-f9688c74a3bb" />

Buscar partido por id
<img width="468" alt="image" src="https://github.com/user-attachments/assets/6a3a8d14-5638-41e7-ad36-106cff5c4a99" />

Eliminar Partido

<img width="468" alt="image" src="https://github.com/user-attachments/assets/2e283fba-0e19-482a-b77f-a1879fbf5cf4" />



Este proyecto implementa un backend en Go que gestiona partidos de La Liga. 
## Endpoints

- **GET /api/matches**: Lista todos los partidos (incluye goles, tarjetas y tiempo extra).
- **GET /api/matches/{id}**: Obtiene los detalles de un partido por ID.
- **POST /api/matches**: Crea un nuevo partido (los campos adicionales se inicializan en 0).
- **PUT /api/matches/{id}**: Actualiza los datos básicos de un partido.
- **DELETE /api/matches/{id}**: Elimina un partido.
- **PATCH /api/matches/{id}/goals**: Actualiza los goles del partido. Se espera un JSON con "homeGoals" y "awayGoals".
- **PATCH /api/matches/{id}/yellowcards**: Registra tarjetas amarillas. Se espera un JSON con "yellowCards".
- **PATCH /api/matches/{id}/redcards**: Registra tarjetas rojas. Se espera un JSON con "redCards".
- **PATCH /api/matches/{id}/extratime**: Registra el tiempo extra. Se espera un JSON con "extraTime".

## Documentación

- Consulta el archivo `swagger.yaml` para ver la documentación completa de la API.
- Link a Postman https://javier-8000794.postman.co/workspace/Javier's-Workspace~11f63a9e-ca72-417d-800a-f33cda6dbae7/request/43575796-68a480db-4943-487c-a651-e0be6c602354?action=share&creator=43575796&ctx=documentation
