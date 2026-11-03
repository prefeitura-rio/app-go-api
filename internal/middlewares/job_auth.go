package middlewares

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/handlers/v1/response"
)

// jobCursoLoaderFunc resolves the course a job belongs to (curso_id of its
// metadata). cursoID is 0 when the job names no course.
type jobCursoLoaderFunc func(ctx context.Context, jobID uuid.UUID) (cursoID int, found bool, err error)

// JobAccess authorizes GET /jobs/:jobId/status with the same rule as the route
// that creates the job (POST /courses/:courseId/enrollments/import, guarded by
// CourseAuthorization + CourseOwnershipCheck), applied to the job's course:
//
//   - admin / go:cursos:casa_civil: any job whose course exists;
//   - go:cursos:editor with secretaria/orgao scope: only jobs of a course in that scope;
//   - anyone else: 403, before the job is looked up, so its existence is not revealed.
//
// A job that names no course is readable only by admin / go:cursos:casa_civil.
func JobAccess(jobLoader jobCursoLoaderFunc, courseLoader courseLoaderFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		unrestricted := IsAdmin(c) || HasRole(c, "go:cursos:casa_civil")
		if !unrestricted && !(isCursosEditor(c) && hasCourseOrgaoScope(c)) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para acessar este recurso"})
			c.Abort()
			return
		}

		jobID, err := uuid.Parse(c.Param("jobId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID do job inválido"})
			c.Abort()
			return
		}

		cursoID, found, err := jobLoader(c.Request.Context(), jobID)
		if err != nil {
			response.FromError(c, err, "Erro ao buscar job")
			c.Abort()
			return
		}
		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job não encontrado"})
			c.Abort()
			return
		}

		if cursoID <= 0 {
			if unrestricted {
				c.Next()
				return
			}
			c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para acessar este recurso"})
			c.Abort()
			return
		}

		authorizeCourseOwnership(c, courseLoader, cursoID)
	}
}
