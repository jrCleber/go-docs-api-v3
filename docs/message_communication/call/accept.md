# Aceita

## `call.accept`

Este evento é acionado no exato momento em que uma chamada feita através do WhatsApp é aceita pelo destinatário. Ele serve como uma notificação imediata que informa tanto o sistema quanto o usuário que a conexão de chamada foi estabelecida com sucesso, permitindo que a comunicação de voz ou vídeo prossiga.

A captura deste evento é crucial para aplicativos que necessitam iniciar ou preparar recursos assim que a chamada é aceita, como habilitar gravações, ativar indicadores de status de chamada em tempo real ou mesmo para análises de dados de uso de chamadas. Além disso, pode ser utilizado para melhorar a experiência do usuário, fornecendo feedback visual ou auditivo que confirme o estabelecimento da chamada.

### Request

#### HEADER PARAMTERS

- `x-instance`: _(string)_ Identificador da inatância;
- `x-whatsapp`: _(string)_ Número do dispositivo conectado.

```json
POST /caminho/recurso HTTP/1.1
Host: www.exemplo.com
Content-Type: application/json
Content-Length: <ComprimentoDosDadosEmBytes>

{
  "chave": "valor",
  "outraChave": "outroValor"
}

```

#### REQUEST BODY: `multipart/form-data`

- `event`: _(string)_ - `"call.accept"`
- `instance`: _(object)_
  - `instanceId`: _(string)_ ID da instância connectdata.
  - `name`: _(string)_ Nome da instânca connectada.
  - `state`: _(enum)_ Indica a condição geral da instância.
    - `active`: _(string)_ indica que a instância está em operação;
    - `inactive`: _(string)_ indica que a instância está desligada.
  - `status`: _(enum)_ Fornece detalhes sobre a fase atual no ciclo de vida da instância. O status pode indicar uma fase transitória, condições temporárias e definitivos
    - `created`: _(string)_ a instância foi criada, mas ainda não está completamente disponível;
    - `waiting`: _(string)_ indica que a instância já está operacional e pronta para a conexão;
    - `available`: _(string)_ indica que a instância já completou todo o processo de conexão e está pronta para uso;
    - `deleted`:  _(string)_ a instância foi desativada e deletada permanentemente.
      - A deleção só irá ocorre após o logout da conexão.
  - `connection`: _(enum)_ Representa o estado atual da conexão com o WhatsApp.
    - `close`: _(string)_
    - `open`: _(string)_
    - `refused`: _(string)_
  - `createdAt`: _(date-time)_ Data da criação da instância.
