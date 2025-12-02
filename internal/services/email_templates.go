package services

import (
	"fmt"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// EmailTemplate defines the structure for email content
type EmailTemplate struct {
	Subject string
	Body    string
	IsHTML  bool
}

// GetEnrollmentPendingEmailTemplate returns email template for "Em Análise" status
func GetEnrollmentPendingEmailTemplate(inscricao *models.Inscricao, curso *models.Curso, prefrioDomain string) EmailTemplate {
	subject := fmt.Sprintf("Inscrição recebida! - %s", curso.Titulo)

	meusCoursosURL := fmt.Sprintf("https://%s/servicos/cursos/meus-cursos", prefrioDomain)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Sua inscrição foi recebida com sucesso e agora está em análise.</p>

    <p>Estamos avaliando as solicitações e entraremos em contato em breve para informar o resultado. Você pode acompanhar o status da sua inscrição a qualquer momento na seção <a href="%s" style="color: #0066cc; text-decoration: none;">Meus Cursos</a> na plataforma do Oportunidades Cariocas.</p>

    <p>Até logo,<br>
    Equipe do Oportunidades Cariocas</p>
</body>
</html>`,
		inscricao.Name,
		meusCoursosURL,
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// GetEnrollmentApprovedEmailTemplate returns email template for "Inscrito" status
func GetEnrollmentApprovedEmailTemplate(inscricao *models.Inscricao, curso *models.Curso, prefrioDomain string) EmailTemplate {
	subject := fmt.Sprintf("Parabéns! Sua inscrição foi aprovada - %s", curso.Titulo)

	meusCoursosURL := fmt.Sprintf("https://%s/servicos/cursos/meus-cursos", prefrioDomain)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Temos uma ótima notícia! Sua inscrição foi aprovada.</p>

    <p>Você está confirmado(a) na atividade. Mantenha-se atento(a) ao seu e-mail e/ou telefone, caso informações adicionais sejam necessárias para a atividade.</p>

    <p>Você pode acompanhar o status da sua inscrição a qualquer momento na seção <a href="%s" style="color: #0066cc; text-decoration: none;">Meus Cursos</a> na plataforma do Oportunidades Cariocas.</p>

    <p>Estamos felizes por ter você conosco!</p>

    <p>Até logo,<br>
    Equipe do Oportunidades Cariocas</p>
</body>
</html>`,
		inscricao.Name,
		meusCoursosURL,
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}

// GetEnrollmentRejectedEmailTemplate returns email template for "Recusado" status
func GetEnrollmentRejectedEmailTemplate(inscricao *models.Inscricao, curso *models.Curso, prefrioDomain string) EmailTemplate {
	subject := fmt.Sprintf("Informações sobre sua inscrição - %s", curso.Titulo)

	meusCoursosURL := fmt.Sprintf("https://%s/servicos/cursos/meus-cursos", prefrioDomain)
	cursosURL := fmt.Sprintf("https://%s/servicos/cursos", prefrioDomain)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <p>Olá, %s!</p>

    <p>Agradecemos o seu interesse na atividade "%s".</p>

    <p>Analisamos sua inscrição, mas, infelizmente, ela não foi aprovada desta vez. A recusa pode ter ocorrido por não ter sido cumprido algum dos requisitos ou critérios definidos pela unidade responsável.</p>

    <p>Você pode acompanhar o status da sua inscrição a qualquer momento na seção <a href="%s" style="color: #0066cc; text-decoration: none;">Meus Cursos</a> na plataforma do Oportunidades Cariocas.</p>

    <p>Não desanime! Convidamos você a conhecer <a href="%s" style="color: #0066cc; text-decoration: none;">outras oportunidades</a> disponíveis em nossa plataforma que podem ser do seu interesse.</p>

    <p>Até logo,<br>
    Equipe do Oportunidades Cariocas</p>
</body>
</html>`,
		inscricao.Name,
		curso.Titulo,
		meusCoursosURL,
		cursosURL,
	)

	return EmailTemplate{
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}
}
