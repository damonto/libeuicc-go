#include <stdint.h>

#include <libqmi-glib.h>

struct qmi_data
{
    int lastChannelId;
    int uimSlot;
    GMainContext *context;
    QmiClientUim *uimClient;
};

QmiDevice *qmi_device_new_from_path(GFile *file, GMainContext *context, GError **error);

gboolean
qmi_device_open_sync(
    QmiDevice *device,
    GMainContext *context,
    GError **error);

QmiClient *
qmi_device_allocate_client_sync(
    QmiDevice *device,
    GMainContext *context,
    GError **error);

gboolean
qmi_device_release_client_sync(
    QmiDevice *device,
    QmiClient *client,
    GMainContext *context,
    GError **error);

QmiMessageUimOpenLogicalChannelOutput *
qmi_client_uim_open_logical_channel_sync(
    QmiClientUim *client,
    QmiMessageUimOpenLogicalChannelInput *input,
    GMainContext *context,
    GError **error);

QmiMessageUimLogicalChannelOutput *
qmi_client_uim_logical_channel_sync(
    QmiClientUim *client,
    QmiMessageUimLogicalChannelInput *input,
    GMainContext *context,
    GError **error);

QmiMessageUimSendApduOutput *
qmi_client_uim_send_apdu_sync(
    QmiClientUim *client,
    QmiMessageUimSendApduInput *input,
    GMainContext *context,
    GError **error);

int libeuicc_qmi_apdu_connect(int uim_slot, char *device_path);
void libeuicc_qmi_apdu_disconnect();
int libeuicc_qmi_apdu_transmit(uint8_t **rx, uint32_t *rx_len, const uint8_t *tx, uint32_t tx_len);
int libeuicc_qmi_apdu_open_logical_channel(const uint8_t *aid, uint8_t aid_len);
int libeuicc_qmi_apdu_close_logical_channel(uint8_t channel);
